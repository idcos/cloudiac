package policy

import (
	"bufio"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/runner"
	"cloudiac/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/hcl"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Parser struct {
}

type Policy struct {
	Id   string `json:"Id"`
	Meta Meta   `json:"meta"`
	Rego string `json:"rego"`
}

type Meta struct {
	Category     string `json:"category"`
	File         string `json:"file"`
	Id           string `json:"id"`
	Name         string `json:"name"`
	PolicyType   string `json:"policy_type"`
	ReferenceId  string `json:"reference_id"`
	ResourceType string `json:"resource_type"`
	Severity     string `json:"severity"`
	Version      int    `json:"version"`
}

type Resource struct {
	ResourceType string `json:"resourceType" enums:"local,remote"`
	RepoAddr     string // repo 远程地址或者本地目录路径
	Token        string
	Revision     string
	SubDir       string

	StopOnViolation bool
	workingDir      string
	codeDir         string
}

type OutputResult struct {
	Results Results `json:"results"`
}

type Results struct {
	PassedRules []Rule      `json:"passed_rules"`
	Violations  []Violation `json:"violations"`
	Count       TsCount     `json:"count"`
}

type Rule struct {
	RuleName    string `json:"rule_name"`
	Description string `json:"description"`
	RuleId      string `json:"rule_id"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
}

type Violation struct {
	RuleName     string `json:"rule_name"`
	Description  string `json:"description"`
	RuleId       string `json:"rule_id"`
	Severity     string `json:"severity"`
	Category     string `json:"category"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	File         string `json:"file"`
	Line         int    `json:"line"`
}

type TsCount struct {
	Low    int `json:"low"`
	Medium int `json:"medium"`
	High   int `json:"high"`
	Total  int `json:"total"`
}

func UnmarshalOutputResult(bs []byte) (*OutputResult, error) {
	js := OutputResult{}
	err := json.Unmarshal(bs, &js)
	return &js, err
}

func (s *Scanner) GetResultPath(res Resource) string {
	return filepath.Join(s.WorkingDir, runner.TerrascanResultFile)
}

func (s *Scanner) GetLogPath() string {
	return filepath.Join(s.WorkingDir, runner.TerrascanLogFile)
}

func (s *Scanner) GetConfigPath(res Resource) string {
	return filepath.Join(s.WorkingDir, runner.TerrascanJsonFile)
}

func (r Resource) GetUrl(task *models.Task) string {
	u := getGitUrl(task.RepoAddr, "", task.CommitId, task.Workdir)
	return u
}

type Scanner struct {
	Db         *db.Session
	Logfp      *os.File
	DebugLog   bool // 是否输出详细调试日志
	ParseOnly  bool // 是否只解析模板
	SaveResult bool // 是否保持扫描结果到数据库（进针对环境和模板扫描）
	RemoteScan bool // 是否由 terrascan 执行远程扫描
	WorkingDir string
	PolicyDir  string

	Policies []Policy

	Resources []Resource

	Result []services.TsResultJson
}

func (p Parser) Parse(filePath string) error {
	//logrus.Errorf("parse \n")

	byt, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	//logrus.Errorf("byt %s\n", byt)

	a, err := hcl.ParseBytes(byt)
	if err != nil {
		return err
	}
	//logrus.Errorf("ast %s\n", a)

	js, err := json.Marshal(a)
	if err != nil {
		return err
	}
	fmt.Printf("%s", js)
	return nil
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
	for _, resource := range s.Resources {
		err = s.ScanResource(resource)
		if err != nil {
			err = fmt.Errorf("scan error %v", err)
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

	return err
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
	task := models.Task{
		OrgId:     "",
		ProjectId: "",
		TplId:     "",
		EnvId:     "",
	}
	task.Id = ""

	if s.SaveResult {
		if err := services.InitScanResult(s.Db, task); err != nil {
			return err
		}
	}

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

	if err := s.RunScan(resource); err != nil {
		code, er := utils.CmdGetCode(err)
		switch code {
		case 3: // found violation, continue process
		case 0, 1:
			fallthrough
		default:
			return s.handleScanError(&task, er)
		}
	}

	bs, err := os.ReadFile(s.GetResultPath(resource))
	if err != nil {
		return s.handleScanError(&task, err)
	}

	var tfResultJson *services.TsResultJson
	if tfResultJson, err = services.UnmarshalTfResultJson(bs); err != nil {
		return s.handleScanError(&task, err)
	}

	if len(tfResultJson.Results.Violations) > 0 {
		// 附加源码
		if tfResultJson, err = populateViolateSource(s, resource, &task, tfResultJson); err != nil {
			return s.handleScanError(&task, err)
		}
	}

	if s.SaveResult {
		if err := services.SaveTfScanResult(s.Db, &task, tfResultJson.Results); err != nil {
			return s.handleScanError(&task, err)
		}

	}

	return nil
}

func populateViolateSource(scanner *Scanner, res Resource, task *models.Task, resultJson *services.TsResultJson) (*services.TsResultJson, error) {
	updated := false
	for idx, policyResult := range resultJson.Results.Violations {
		fmt.Printf("violation %+v", policyResult)
		if policyResult.File == "" {
			continue
		}

		srcFile, err := os.Open(filepath.Join(scanner.WorkingDir, res.codeDir, policyResult.File))
		if err != nil {
			fmt.Printf("open err %+v", err)

			continue
		}
		reader := bufio.NewReader(srcFile)
		for lineNo := 1; ; lineNo++ {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				break
			}
			if lineNo < policyResult.Line {
				continue
			} else if lineNo-policyResult.Line >= runner.PopulateSourceLineCount {
				break
			}

			resultJson.Results.Violations[idx].Source += string(line)
			updated = true
		}
		fmt.Printf("violation with src %+v", resultJson.Results.Violations[idx])
		_ = srcFile.Close()
	}
	fmt.Printf("updaetd %s", updated)

	if updated {
		if js, err := json.MarshalIndent(resultJson, "", "  "); err == nil {
			fmt.Printf("result file %s", scanner.GetResultPath(res))
			err := os.WriteFile(scanner.GetResultPath(res), js, 0644)
			if err != nil {
				fmt.Printf("update result error %+v", err)
			}
		}

	}

	return resultJson, nil
}

func (s *Scanner) handleScanError(task *models.Task, err error) error {
	if s.SaveResult {
		// 扫描出错的时候更新所有策略扫描结果为 failed
		emptyResult := services.TsResultJson{}
		if err := services.SaveTfScanResult(s.Db, task, emptyResult.Results); err != nil {
			return err
		}
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

-o json > {{.TerrascanResultFile}} 2>{{.TerrascanLogFile}}
`
	cmdline := utils.SprintTemplate(cmdlineTemplate, map[string]interface{}{
		"CodeDir":             s.WorkingDir,
		"TerrascanResultFile": s.GetResultPath(res),
		"TerrascanLogFile":    s.GetLogPath(),
		"PolicyDir":           s.PolicyDir,
		"CloudIacDebug":       s.DebugLog,
	})
	return cmdline
}

var scanInitCommandTpl = `#!/bin/sh
git clone '{{.RepoAddress}}' code && \
cd 'code/{{.SubDir}}' && \
git checkout -q '{{.Revision}}' && echo check out $(git rev-parse --short HEAD). 
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

func NewScanner(resources []Resource) (*Scanner, error) {
	scanner := Scanner{
		Resources: resources,
	}
	return &scanner, nil
}

// genPolicyFiles 将策略文件写入策略目录
func genPolicyFiles(policyDir string, policies []Policy) error {
	if err := os.MkdirAll(policyDir, 0755); err != nil {
		return err
	}

	for _, policy := range policies {
		if err := os.MkdirAll(filepath.Join(policyDir, policy.Id), 0755); err != nil {
			return err
		}
		js, _ := json.Marshal(policy.Meta)
		if err := os.WriteFile(filepath.Join(policyDir, policy.Id, "meta.json"), js, 0644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(policyDir, policy.Id, "policy.rego"), []byte(policy.Rego), 0644); err != nil {
			return err
		}
	}
	return nil
}

// getGitUrl 获取 go-getter git 下载地址
// version 可以是 branch、tag 名称 或者 commit hash
// 同时可以通过 subDir 指定只下载某个子目录
// example: git::http://token:the_token@git.example.com/git_sample/repo-example.git//dev?ref=v1.0.0
func getGitUrl(repoAddr, token, version, subDir string) string {
	u, err := url.Parse(repoAddr)
	if err != nil {
		return ""
	}

	// gitlab http token 认证
	if token != "" {
		u.User = url.UserPassword("token", token)
	}

	query := url.Values{}
	if version != "" {
		query.Add("ref", version)
	}
	u.RawQuery = query.Encode()

	// get-getter 子目录使用双斜杠
	if subDir != "" {
		if subDir[0:1] != "/" {
			subDir = "/" + subDir
		}
		subDir = "/" + subDir
		u.Path = u.Path + subDir
	}

	return u.String()
}

func GetPoliciesFromDB(query *db.Session, policyIds []string) ([]Policy, error) {
	var policies []models.Policy
	if err := query.Model(models.Policy{}).Where("id in (?)").Find(&policies); err != nil {
		return nil, err
	}
	if len(policies) != len(policyIds) {
		return nil, fmt.Errorf("invalid policy id found")
	}

	var retPolicies []Policy
	for _, policy := range policies {
		category := "general"
		group, _ := services.GetPolicyGroupById(query, policy.GroupId)
		if group != nil {
			category = group.Name
		}
		retPolicies = append(retPolicies, Policy{
			Id: string(policy.Id),
			Meta: Meta{
				Category:     category,
				File:         "policy.rego",
				Id:           string(policy.Id),
				Name:         policy.Entry,
				PolicyType:   policy.PolicyType,
				ReferenceId:  policy.ReferenceId,
				ResourceType: policy.ResourceType,
				Severity:     policy.Severity,
				Version:      policy.Revision,
			},
			Rego: policy.Rego,
		})
	}
	return retPolicies, nil
}

func NewScannerFromLocalDir(srcPath string, policyDir string) (*Scanner, error) {
	res := Resource{
		ResourceType: "local",
		RepoAddr:     srcPath,
	}
	scanner, err := NewScanner([]Resource{res})
	scanner.PolicyDir = policyDir

	return scanner, err
}

func NewScannerFromEnv(query *db.Session, envId string) (*Scanner, error) {
	// 获取 git 仓库地址
	env, err := services.GetEnvById(query, models.Id(envId))
	if err != nil {
		return nil, err
	}
	tpl, err := services.GetTemplateById(query, env.TplId)
	if err != nil {
		return nil, err
	}
	repoAddr, commitId, err := services.GetTaskRepoAddrAndCommitId(query, tpl, env.Revision)

	if err != nil {
		return nil, err
	}
	res := Resource{
		ResourceType: "remote",
		RepoAddr:     repoAddr,
		Revision:     commitId,
		SubDir:       tpl.Workdir,
	}

	// 获取 环境 关联 策略组 和 策略列表
	policies, err := services.GetPoliciesByEnvId(query, env.Id)
	if err != nil {
		return nil, err
	}
	if len(policies) == 0 {
		return nil, fmt.Errorf("no valid policy found")
	}

	scanner, er := NewScanner([]Resource{res})
	if er != nil {
		return nil, er
	}

	if policies, err := GetScanPolicies(query, policies); err != nil {
		return nil, err
	} else {
		scanner.Policies = policies
	}

	return scanner, nil
}

func NewScannerFromTemplate(query *db.Session, tplId string) (*Scanner, error) {
	// 获取 git 仓库地址
	tpl, err := services.GetTemplateById(query, models.Id(tplId))
	if err != nil {
		return nil, err
	}
	repoAddr, commitId, err := services.GetTaskRepoAddrAndCommitId(query, tpl, tpl.RepoRevision)
	if err != nil {
		return nil, err
	}
	res := Resource{
		ResourceType: "remote",
		RepoAddr:     repoAddr,
		Revision:     commitId,
		SubDir:       tpl.Workdir,
	}

	// 获取 模板 关联 策略组 和 策略列表
	policies, err := services.GetPoliciesByTemplateId(query, tpl.Id)
	if err != nil {
		return nil, err
	}
	if len(policies) == 0 {
		return nil, fmt.Errorf("no valid policy found")
	}
	scanner, er := NewScanner([]Resource{res})
	if er != nil {
		return nil, er
	}

	if policies, err := GetScanPolicies(query, policies); err != nil {
		return nil, err
	} else {
		scanner.Policies = policies
	}

	return scanner, nil
}

func GetScanPolicies(query *db.Session, policies []models.Policy) ([]Policy, error) {
	var ps []Policy
	for _, policy := range policies {
		group, err := services.GetPolicyGroupById(query, policy.GroupId)
		if err != nil {
			return nil, err
		}
		ps = append(ps, Policy{
			Id: string(policy.Id),
			Meta: Meta{
				Category:     group.Name,
				File:         "policy.rego",
				Id:           string(policy.Id),
				Name:         policy.Name,
				PolicyType:   policy.PolicyType,
				ReferenceId:  policy.ReferenceId,
				ResourceType: policy.ResourceType,
				Severity:     policy.Severity,
				Version:      policy.Revision,
			},
			Rego: policy.Rego,
		})
	}

	return ps, nil
}
