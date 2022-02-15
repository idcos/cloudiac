// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package policy

import (
	"cloudiac/common"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/runner"
	"cloudiac/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	ErrScanExitViolated = errors.New("scan completed with violation")
	ErrScanExitFailed   = errors.New("scan failed")
)

type Scanner struct {
	Db         *db.Session
	Logfp      *os.File
	DebugLog   bool   // 是否输出详细调试日志
	ParseOnly  bool   // 是否只解析模板
	SaveResult bool   // 是否保持扫描结果到数据库（进针对环境和模板扫描）
	RemoteScan bool   // 是否由 terrascan 执行远程扫描
	Internal   bool   // 是否使用内置引擎执行扫描
	InputFile  string // 资源输入文件
	ResultFile string // 扫描结果输出，默认输出到 stdout
	MapFile    string // 源码映射文件
	WorkingDir string
	PolicyDir  string

	Policies []Policy

	Resources []Resource

	Result []TsResultJson
}

func (s *Scanner) GetResultPath(res Resource) string {
	if s.ResultFile != "" {
		return filepath.Join(s.WorkingDir, s.ResultFile)
	}
	return filepath.Join(s.WorkingDir, runner.ScanResultFile)
}

func (s *Scanner) GetLogPath() string {
	return filepath.Join(s.WorkingDir, runner.ScanLogFile)
}

func (s *Scanner) GetConfigPath(res Resource) string {
	if res.InputFile != "" {
		return filepath.Join(s.WorkingDir, res.InputFile)
	}
	return filepath.Join(s.WorkingDir, runner.ScanInputFile)
}

func (s *Scanner) Run() error {
	var (
		err error
	)

	if s.SaveResult {
		s.Db = s.Db.Begin()

		defer func() {
			if r := recover(); r != nil {
				_ = s.Db.Rollback()
				panic(r)
			}
		}()
	}

	if err = s.Prepare(); err != nil {
		return err
	}

	// 批量扫描仓库
	var errExit error
	for _, resource := range s.Resources {
		errExit = s.ScanResource(resource)
		if errExit != nil && !errors.Is(errExit, ErrScanExitViolated) {
			break // break to cleanup
		}
	}

	if err = s.CleanUp(err); err != nil {
		return err
	}

	if s.SaveResult {
		if err != nil {
			_ = s.Db.Rollback()
			return e.New(e.DBError, err)
		} else {
			if err = s.Db.Commit(); err != nil {
				_ = s.Db.Rollback()
				return e.New(e.DBError, err)
			}
		}
	}

	return errExit
}

func (s *Scanner) Prepare() error {
	var (
		err error
	)

	// TODO
	s.WorkingDir = "."

	if len(s.Policies) > 0 {
		s.PolicyDir = filepath.Join(s.WorkingDir, runner.PoliciesDir)
		if s.ParseOnly {
			if err := os.MkdirAll(s.PolicyDir, 0755); err != nil {
				return err
			}
		} else {
			if err := genPolicyFiles(s.PolicyDir, s.Policies); err != nil {
				return err
			}
		}
	}

	s.Logfp, err = os.OpenFile(s.GetLogPath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	// 创建 terrascan 默认策略目录避免网络请求
	homeDir, _ := homedir.Dir()
	_ = os.MkdirAll(filepath.Join(homeDir, ".terrascan/pkg/policies/opa/rego/aws"), 0755)

	return nil
}

// CleanUp 清理
func (s *Scanner) CleanUp(er error) error {
	//err := os.RemoveAll(s.WorkingDir)
	//if err != nil {
	//	return err
	//}

	if err := s.Logfp.Close(); err != nil {
		return err
	}

	return nil
}

func (s *Scanner) ScanResource(resource Resource) error {
	// TODO: handle relation with task
	task := models.ScanTask{
		OrgId:     "",
		ProjectId: "",
		TplId:     "",
		EnvId:     "",
	}
	task.Id = ""

	//if s.SaveResult {
	//	if err := services.InitScanResult(s.Db, &task); err != nil {
	//		return err
	//	}
	//}

	if resource.ResourceType == "remote" && !s.RemoteScan {
		resource.codeDir = "code"
		cmdline := s.genScanInit(&resource)
		cmd := exec.Command("sh", "-c", cmdline)
		output, err := cmd.CombinedOutput()
		s.Logfp.Write(output)
		if err != nil {
			return fmt.Errorf("checkout error %v, output %s", err, output)
		}
	}

	var errExit error
	if s.Internal {
		if errExit = s.RunInternalScan(resource); errExit != nil && !errors.Is(errExit, ErrScanExitViolated) {
			task.PolicyStatus = common.PolicyStatusFailed
			return errExit
		}
	} else {
		if err := s.RunScan(resource); err != nil {
			defer s.handleScanError(&task, err)
			code, err := utils.CmdGetCode(err)
			if err != nil {
				return err
			}
			switch code {
			case 3:
				task.PolicyStatus = common.PolicyStatusViolated
			case 0:
				task.PolicyStatus = common.PolicyStatusPassed
			case 1:
				task.PolicyStatus = common.PolicyStatusFailed
				return err
			default:
				task.PolicyStatus = common.PolicyStatusFailed
				return err
			}
		}
	}

	bs, err := os.ReadFile(s.GetResultPath(resource))
	if err != nil {
		return err
	}

	var tfResultJson *TsResultJson
	if tfResultJson, err = UnmarshalTfResultJson(bs); err != nil {
		return err
	}

	if len(tfResultJson.Results.Violations) > 0 {
		// 附加源码
		if tfResultJson, err = PopulateViolateSource(s, resource, &task, tfResultJson); err != nil {
			return err
		}
	}

	//if s.SaveResult {
	//	if err := services.UpdateScanResult(s.Db, &task, tfResultJson.Results, task.PolicyStatus); err != nil {
	//		return err
	//	}
	//}

	return errExit
}

func (s *Scanner) GetMessage(format string, data interface{}) string {
	return utils.SprintTemplate(format, data)
}

func (s *Scanner) handleScanError(task *models.ScanTask, err error) error {
	if s.SaveResult {
		// 扫描出错的时候更新所有策略扫描结果为 failed
		//emptyResult := TsResultJson{}
		//if err := services.UpdateScanResult(s.Db, task, emptyResult.Results, task.PolicyStatus); err != nil {
		//	return err
		//}
	}

	return err
}

// TODO
func (s *Scanner) genScanScript(res Resource) string {
	cmdlineTemplate := `
cd {{.CodeDir}} && \
mkdir -p ~/.terrascan/pkg/policies/opa/rego && \
terrascan scan -d . -p {{.PolicyDir}} --show-passed \
{{if .TfVars}}-var-file={{.TfVars}}{{end}} \

-o json > {{.ScanResultFile}} 2>{{.ScanLogFile}}
`
	cmdline := utils.SprintTemplate(cmdlineTemplate, map[string]interface{}{
		"CodeDir":        s.WorkingDir,
		"ScanResultFile": s.GetResultPath(res),
		"ScanLogFile":    s.GetLogPath(),
		"PolicyDir":      s.PolicyDir,
		"CloudIacDebug":  s.DebugLog,
	})
	return cmdline
}

var scanInitCommandTpl = `#!/bin/sh
git clone '{{.RepoAddress}}' code && \
cd code && \
echo "checkout $(git rev-parse --short HEAD)." && \
git checkout -q '{{.Revision}}' && \
cd '{{.SubDir}}'
`

func (s *Scanner) genScanInit(res *Resource) (command string) {
	cmdline := utils.SprintTemplate(scanInitCommandTpl, map[string]interface{}{
		"RepoAddress": res.RepoAddr,
		"SubDir":      res.SubDir,
		"Revision":    res.Revision,
	})

	return cmdline
}

func (s *Scanner) RunScan(resource Resource) error {
	var (
		cmdErr  error
		timeout = 30 * time.Second
	)

	cmd := exec.Command("terrascan", "scan")
	cmd.Dir = s.WorkingDir

	// 扫描目标，可以通过 terrascan 远程扫描或者手动 checkout 后扫描本地
	if resource.ResourceType == "remote" && s.RemoteScan {
		address := getGitUrl(resource.RepoAddr, resource.codeDir, resource.Revision, resource.SubDir)
		if address == "" {
			return errors.New("get git url error")
		}
		cmd.Args = append(cmd.Args, "-r", "git", "-u", address)
	} else {
		// 本地扫描
		cmd.Args = append(cmd.Args, "-d", resource.codeDir)
	}

	// 策略目录
	cmd.Args = append(cmd.Args, "-p", s.PolicyDir)
	// 输出结果为 json 格式
	cmd.Args = append(cmd.Args, "-o", "json")
	// 结果包含已经通过的规则
	cmd.Args = append(cmd.Args, "--show-passed")
	// 包含完整调试日志
	if s.DebugLog {
		cmd.Args = append(cmd.Args, "-l", "debug")
	}
	resultFile := s.GetResultPath(resource)
	// 只解析，输出到 _tsscan.json
	if s.ParseOnly {
		cmd.Args = append(cmd.Args, "--config-only")
		resultFile = s.GetConfigPath(resource)
	}
	resultFp, err := os.OpenFile(resultFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	cmd.Stdout = resultFp
	cmd.Stderr = s.Logfp

	done := make(chan error)
	go func() {
		if err := cmd.Start(); err != nil {
			logrus.Errorf("error start cmd %s, err: %v", cmd.Path, err)
			done <- err
		}
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		if err := cmd.Process.Kill(); err != nil {
			logrus.Errorf("kill timeout process %s error %v", cmd.Path, err)
		} else {
			logrus.Errorf("kill timeout process %s with timeout %s seconds", cmd.Path, timeout)
		}
		cmdErr = fmt.Errorf("process timeout")
	case err = <-done:
		logrus.Errorf("command complete with error %v", err)
		cmdErr = err
	}

	return cmdErr
}

func (s *Scanner) RunInternalScan(code Resource) error {
	output := TsResultJson{}
	output.Results.ScanSummary.ScannedAt = fmt.Sprintf("%s", time.Now().Format(time.RFC3339))
	output.Results.ScanSummary.IacType = "terraform"
	output.Results.ScanSummary.FileFolder = code.codeDir

	policies, err := s.ReadPolicies(s.PolicyDir)
	if err != nil {
		return err
	}

	inputResource := models.TfParse{}
	inputContent, _ := ioutil.ReadFile(s.GetConfigPath(code))
	if len(inputContent) > 0 {
		_ = json.Unmarshal(inputContent, &inputResource)
	}

	violated := false
	for _, p := range policies {
		result, err := RegoParse(filepath.Join(p.Meta.Root, p.Meta.File), s.GetConfigPath(code), p.Meta.Name)
		if err != nil {
			scanError := ScanError{
				RuleName:    p.Meta.Name,
				Description: p.Meta.Description,
				Severity:    p.Meta.Severity,
				Category:    p.Meta.Category,
				IacType:     "terraform",
				Directory:   "code",
				RuleId:      p.Meta.Id,
				File:        p.Meta.File,
				Error:       err,
			}
			output.Results.ScanErrors = append(output.Results.ScanErrors, scanError)
			output.Results.ScanSummary.PoliciesError++
			s.Console(s.GetMessage(MSG_TEMPLATE_ERROR, scanError))
			continue
		}
		// parse result
		res := (&Rego{}).ParseResource(result)
		// generate result
		if len(res) > 0 {
			resName := res[0]
			resType := res[0][0:strings.Index(res[0], ".")]
			violation := Violation{
				RuleName:     p.Meta.Name,
				Description:  p.Meta.Description,
				RuleId:       p.Meta.Id,
				Severity:     p.Meta.Severity,
				Category:     p.Meta.Category,
				ResourceName: resName,
				ResourceType: resType,
			}
			if len(inputResource) > 0 {
				violation.Line, violation.File = findLineNoFromMap(inputResource, resName)
			}
			output.Results.Violations = append(output.Results.Violations, violation)
			output.Results.ScanSummary.ViolatedPolicies++
			s.Console(s.GetMessage(MSG_TEMPLATE_VIOLATED, violation))
			violated = true
		} else {
			rule := Rule{
				RuleName:    p.Meta.Name,
				Description: p.Meta.Description,
				RuleId:      p.Meta.Id,
				Severity:    p.Meta.Severity,
				Category:    p.Meta.Category,
			}
			output.Results.PassedRules = append(output.Results.PassedRules, rule)
			output.Results.ScanSummary.PoliciesValidated++
			s.Console(s.GetMessage(MSG_TEMPLATE_PASSED, rule))
		}
		switch strings.ToLower(p.Meta.Severity) {
		case common.PolicySeverityHigh:
			output.Results.ScanSummary.High++
		case common.PolicySeverityMedium:
			output.Results.ScanSummary.Medium++
		case common.PolicySeverityLow:
			output.Results.ScanSummary.Low++
		default:
			output.Results.ScanSummary.Medium++
		}
	}

	if s.ResultFile != "" {
		outputB, _ := json.Marshal(output)
		if err := os.WriteFile(s.GetResultPath(code), outputB, 0644); err != nil {
			return err
		}
	} else {
		outputB, _ := json.MarshalIndent(output, "", "  ")
		fmt.Printf("%s\n", outputB)
	}

	if violated {
		return ErrScanExitViolated
	}

	return nil
}

func (s *Scanner) ReadPolicies(policyDir string) ([]*PolicyWithMeta, error) {
	// 文件结构：
	// policies
	// └── po-c79sciqs1s43tkfgp2cg
	//     ├── meta.json
	//     └── policy.rego
	policiesDir, err := ioutil.ReadDir(policyDir)
	if err != nil {
		return nil, err
	}

	var ret []*PolicyWithMeta
	// 遍历当前目录
	for _, f := range policiesDir {
		policies, err := ParsePolicyGroup(filepath.Join(policyDir, f.Name()))
		if err != nil {
			scanError := ScanError{
				RuleName: f.Name(),
				File:     f.Name(),
				RuleId:   f.Name(),
				Error:    err,
			}
			s.Console(s.GetMessage(MSG_TEMPLATE_INVALID, scanError))
		}
		ret = append(ret, policies...)
	}
	return ret, nil
}

func (s *Scanner) Console(msg string) {
	fmt.Printf("%s\n", msg)
}

func NewScanner(resources []Resource) (*Scanner, error) {
	scanner := Scanner{
		Resources: resources,
	}
	return &scanner, nil
}
