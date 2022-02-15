// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package main

import (
	"bytes"
	"cloudiac/configs"
	"cloudiac/policy"
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

	"github.com/pkg/errors"

	"github.com/itchyny/gojq"
)

// iac-tool scan 执行策略扫描
//    --debug xxx.tf xxx.rego 调用 opa 引擎执行裸 rego 脚本
//    -d code -p policies 调用 terrascan 扫描本地目录
//    -e envId 调用 terrascan 扫描环境
//    -t tplId 调用 terrascan 扫描模板
//    -r 对环境/模板执行远程扫描
//    --parse 只解析模板不扫描
//    -s 存储扫描结果到数据库
//    -v 详细扫描日志
//
// Example:
// 1. 扫描本地 code 目录（策略使用 policies目录）
//    iac-tool scan
// 2. 扫描环境
//    iac-tool scan -e env-xxxxxx
// 3. 扫描模板
//    iac-tool scan -t tpl-xxxxxx
// 4. 调试 rego 脚本
//    iac-tool scan --debug xxx.tf xxx.rego
// 5. 内置引擎扫描
//    iac-tool scan --internal -p policies -f tfscan.json -o tfscan.json

type ScanCmd struct {
	Debug          bool   `long:"debug" description:"run raw rego script \nuse \"--debug -d code xxx.rego\" or \"--debug xxx.tf xxx.rego\"" required:"false"`
	CodeDir        string `long:"code-dir" short:"d" description:"local source code directory, default:\"code\"" required:"false"`
	PolicyDir      string `long:"policy-dir" short:"p" description:"local policy directory, default:\"policies\"" required:"false"`
	EnvId          string `long:"envId" short:"e" description:"scan environment with envId" required:"false"`
	TplId          string `long:"tplId" short:"t" description:"scan template with tplId" required:"false"`
	ParseOnly      bool   `long:"parse" description:"parse template to json" required:"false"`
	SaveResultToDB bool   `long:"save-result" short:"s" description:"save scan result to database, default:false" required:"false"`
	//PolicyId       string `long:"policy-id" short:"i" description:"scan with policy id, multiple id using \"id1,id2,...\"" required:"false"`
	//PolicyGroupId  string `long:"policy-group-id" short:"g" description:"scan with policy group id, multiple id using \"id1,id2,...\"" required:"false"`
	RemoteScan    bool   `long:"remote-scan" short:"r" description:"scan environment/template remotely" required:"false"`
	Verbose       bool   `long:"verbose" short:"v" description:"write verbose scan log message" required:"false"`
	ParsePlan     bool   `long:"parse-plan" description:"parse tfplan to input.json" required:"false"`
	PlanFile      string `long:"plan" description:"the tfplan json file path" required:"false"`
	JsonFile      string `long:"json" short:"o" description:"the json file path to output, default: output to stdout" required:"false"`
	Internal      bool   `long:"internal" description:"use internal scan engine to execute scan" required:"false"`
	InputFile     string `long:"input" short:"i" description:"the input json file path" required:"false"`
	SourceMapFile string `long:"map" short:"m" description:"the source map json file path" required:"false"`
}

func (*ScanCmd) Usage() string {
	return ""
}

// hasDB 是否需要读取数据库
// 如果在命令行模式执行，并且不需要读取环境/模板/策略，则可以运行在无数据库状态
func (c *ScanCmd) hasDB() bool {
	return c.EnvId != "" || c.TplId != "" ||
		//c.PolicyId != "" || c.PolicyGroupId != "" ||
		c.SaveResultToDB
}

func (c *ScanCmd) Execute(args []string) error {

	if c.Debug {
		filePath := "."
		regoFile := ""
		if c.CodeDir != "" {
			// iac-tool scan --debug -d code xxx.rego
			filePath = c.CodeDir
			if len(args) < 1 {
				return fmt.Errorf("missing iac file or rego script")
			}
			regoFile = args[0]
		} else {
			// iac-tool scan --debug xxx.tf xxx.rego
			if len(args) < 2 {
				return fmt.Errorf("missing iac file or rego script")
			}
			filePath = args[0]
			regoFile = args[1]
		}
		return c.RunDebug(filePath, regoFile)
	}
	if c.ParsePlan {
		return ParseTfplan(c.PlanFile, c.JsonFile)
	}

	if c.hasDB() {
		configs.Init(opt.Config)
		db.Init(configs.Get().Mysql)
		models.Init(false)
	}

	var (
		scanner *policy.Scanner
		er      error
	)

	if c.EnvId != "" {
		return fmt.Errorf("not implement")
		//query := db.Get()
		//scanner, er = NewScannerFromEnv(query, c.EnvId)
		//if er != nil {
		//	return er
		//}
	} else if c.TplId != "" {
		return fmt.Errorf("not implement")
		//query := db.Get()
		//scanner, er = NewScannerFromTemplate(query, c.TplId)
		//if er != nil {
		//	return er
		//}
	} else {
		// 执行本地扫描
		if c.CodeDir == "" {
			c.CodeDir = "code"
		}
		if !utils.FileExist(c.CodeDir) && !c.Internal {
			return fmt.Errorf("missing code dir")
		}
		if c.PolicyDir == "" {
			c.PolicyDir = "policies"
		}
		if !utils.FileExist(c.PolicyDir) {
			return fmt.Errorf("missing policy dir")
		}
		scanner, er = policy.NewScannerFromLocalDir(c.CodeDir, c.PolicyDir, c.InputFile, c.SourceMapFile)
		if er != nil {
			return er
		}
	}

	if c.EnvId != "" || c.TplId != "" {
		scanner.SaveResult = c.SaveResultToDB
	}
	scanner.RemoteScan = c.RemoteScan
	scanner.ParseOnly = c.ParseOnly

	if c.Verbose {
		scanner.DebugLog = c.Verbose
	} else {
		scanner.DebugLog = utils.IsTrueStr(os.Getenv("CLOUDIAC_DEBUG"))
	}
	if c.hasDB() {
		scanner.Db = db.Get()
	}
	if c.Internal {
		scanner.Internal = true
	}
	if c.JsonFile != "" {
		scanner.ResultFile = c.JsonFile
	}
	if c.SourceMapFile != "" {
		scanner.MapFile = c.SourceMapFile
	}

	err := scanner.Run()
	if err != nil {
		if errors.Is(err, policy.ErrScanExitViolated) {
			os.Exit(3)
		} else {
			os.Exit(1)
		}
	}

	return nil
}

func (c *ScanCmd) Parse(filePath string) error {
	cmdString := utils.SprintTemplate("terrascan scan --parse-only -d . -o json > {{.ScanResultFile}}", map[string]interface{}{
		"TFScanJsonFilePath": filepath.Join("./", runner.ScanInputFile),
	})
	result, err := RunCmd(cmdString)
	if err != nil {
		return err
	}
	fmt.Printf("parse result: %s\n", result)
	return nil
}

func (c *ScanCmd) Scan(filePath string, regoDir string) error {
	cmdString := utils.SprintTemplate("terrascan scan -d . -o json > {{.ScanResultFile}}", map[string]interface{}{
		"TFScanJsonFilePath": filepath.Join("./", runner.ScanResultFile),
	})
	result, err := RunCmd(cmdString)
	if err != nil {
		return err
	}
	fmt.Printf("scan result: %s\n", result)
	return nil
}

func (c *ScanCmd) RunDebug(configFile string, regoFile string) error {
	fileInfo, err := os.Stat(configFile)
	if err != nil {
		return err
	}
	isDir := false
	if fileInfo.IsDir() {
		isDir = true
	}

	// 创建空目录，避免 terrascan 读取默认策略
	randomDir, err := os.MkdirTemp(".", "*")
	if err != nil {
		return err
	}

	cmdString := utils.SprintTemplate(`terrascan scan {{if .IsDir}} -d {{.ConfigFile}}{{else}}--iac-file {{.ConfigFile}}{{end}} -p {{.PolicyDir}} --config-only -o json > {{.JsonFile}} && \
opa eval -f pretty --data {{.RegoFile}} --input {{.JsonFile}} data > {{.RegoResultFile}}
`, map[string]interface{}{
		"PolicyDir":      randomDir,
		"ConfigFile":     configFile,
		"IsDir":          isDir,
		"JsonFile":       runner.ScanInputFile,
		"RegoFile":       regoFile,
		"RegoResultFile": runner.RegoResultFile,
	})

	_, err = RunCmd(cmdString)
	if err != nil {
		_ = os.RemoveAll(randomDir)
		return err
	}
	_ = os.RemoveAll(randomDir)
	return nil
}

func RunCmd(cmdString string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", cmdString)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

type ParseCmd struct {
}

func (*ParseCmd) Usage() string {
	return ""
}

func (c *ParseCmd) Execute(args []string) error {
	// iac-tool parse xxx.rego xxx.json
	if len(args) < 2 {
		return fmt.Errorf("missing iac file or rego script")
	}
	regoFile := args[0]
	inputPath := args[1]
	res, err := policy.RegoParse(regoFile, inputPath)
	fmt.Printf("Execute rego parse return res %s, err %+v", res, err)
	return err
}

var (
	changesFilter = `[.resource_changes | .. | select(.type? != null and .address? != null and .mode? == "managed") | {id: .address?, type: .type?, name: .name?, config: (.change.after? + .change.after_unknown?), source: "", line: 0}] | group_by(.type) | map({key:(.[0].type),value:[ .[] ]}) | from_entries`
)

func ParseTfplan(planJsonFile string, planOutputFile string) error {
	if planJsonFile == "" {
		planJsonFile = "tfplan.json"
	}

	tfjson, err := ioutil.ReadFile(planJsonFile)
	if err != nil {
		return err
	}
	tfplan := make(map[string]interface{})
	err = json.Unmarshal(tfjson, &tfplan)
	if err != nil {
		return err
	}
	query, err := gojq.Parse(changesFilter)
	if err != nil {
		return err
	}
	var output []byte
	iter := query.Run(tfplan)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return err
		}

		filtered, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		output = append(output, filtered...)
	}

	if planOutputFile == "" {
		fmt.Printf("%s\n", output)
	} else {
		err = ioutil.WriteFile(planOutputFile, output, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

//
//func NewScannerFromEnv(query *db.Session, envId string) (*policy.Scanner, error) {
//	// 获取 git 仓库地址
//	env, err := services.GetEnvById(query, models.Id(envId))
//	if err != nil {
//		return nil, err
//	}
//	tpl, err := services.GetTemplateById(query, env.TplId)
//	if err != nil {
//		return nil, err
//	}
//	repoAddr, commitId, err := services.GetTaskRepoAddrAndCommitId(query, tpl, env.Revision)
//
//	if err != nil {
//		return nil, err
//	}
//	res := policy.Resource{
//		ResourceType: "remote",
//		RepoAddr:     repoAddr,
//		Revision:     commitId,
//		SubDir:       tpl.Workdir,
//	}
//
//	// 获取 环境 关联 策略组 和 策略列表
//	policies, err := services.GetPoliciesByEnvId(query, env.Id)
//	if err != nil {
//		return nil, err
//	}
//	if len(policies) == 0 {
//		return nil, fmt.Errorf("no valid policy found")
//	}
//
//	scanner, er := policy.NewScanner([]policy.Resource{res})
//	if er != nil {
//		return nil, er
//	}
//
//	if policies, err := services.GetScanPolicies(query, policies); err != nil {
//		return nil, err
//	} else {
//		scanner.Policies = policies
//	}
//
//	return scanner, nil
//}
//
//func NewScannerFromTemplate(query *db.Session, tplId string) (*policy.Scanner, error) {
//	// 获取 git 仓库地址
//	tpl, err := services.GetTemplateById(query, models.Id(tplId))
//	if err != nil {
//		return nil, err
//	}
//	repoAddr, commitId, err := services.GetTaskRepoAddrAndCommitId(query, tpl, tpl.RepoRevision)
//	if err != nil {
//		return nil, err
//	}
//	res := policy.Resource{
//		ResourceType: "remote",
//		RepoAddr:     repoAddr,
//		Revision:     commitId,
//		SubDir:       tpl.Workdir,
//	}
//
//	// 获取 模板 关联 策略组 和 策略列表
//	policies, err := services.GetPoliciesByTemplateId(query, tpl.Id)
//	if err != nil {
//		return nil, err
//	}
//	if len(policies) == 0 {
//		return nil, fmt.Errorf("no valid policy found")
//	}
//	scanner, er := policy.NewScanner([]policy.Resource{res})
//	if er != nil {
//		return nil, er
//	}
//
//	if policies, err := services.GetScanPolicies(query, policies); err != nil {
//		return nil, err
//	} else {
//		scanner.Policies = policies
//	}
//
//	return scanner, nil
//}
