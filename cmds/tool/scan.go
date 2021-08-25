package main

import (
	"bytes"
	"cloudiac/configs"
	"cloudiac/policy"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/runner"
	"cloudiac/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

type ScanCmd struct {
	Debug          bool   `long:"debug" description:"run raw rego script" required:"false"`
	TemplateDir    string `long:"template-dir" short:"d" description:"local source code directory, default:\"code\"" required:"false"`
	PolicyDir      string `long:"policy-dir" short:"p" description:"local policy directory, default:\"policies\"" required:"false"`
	EnvId          string `long:"envId" short:"e" description:"scan environment with envId" required:"false"`
	TplId          string `long:"tplId" short:"t" description:"scan template with tplId" required:"false"`
	ParseOnly      bool   `long:"parse" description:"parse template to json" required:"false"`
	SaveResultToDB bool   `long:"save-result" short:"s" description:"save scan result to database, default:false" required:"false"`
	//PolicyId       string `long:"policy-id" short:"i" description:"scan with policy id, multiple id using \"id1,id2,...\"" required:"false"`
	//PolicyGroupId  string `long:"policy-group-id" short:"g" description:"scan with policy group id, multiple id using \"id1,id2,...\"" required:"false"`
	RemoteScan bool `long:"remote-scan" short:"r" description:"scan environment/template remotely" required:"false"`
	Verbose    bool `long:"verbose" short:"v" description:"write verbose scan log message" required:"false"`
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
		// TODO call opa
		return fmt.Errorf("not implement")
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
		query := db.Get()
		scanner, er = policy.NewScannerFromEnv(query, c.EnvId)
		if er != nil {
			return er
		}
	} else if c.TplId != "" {
		query := db.Get()
		scanner, er = policy.NewScannerFromTemplate(query, c.TplId)
		if er != nil {
			return er
		}
	} else {
		// 执行本地扫描
		if c.TemplateDir == "" {
			c.TemplateDir = "code"
		}
		if c.PolicyDir == "" {
			c.PolicyDir = "policies"
		}
		if !utils.FileExist(c.TemplateDir) || !utils.FileExist(c.PolicyDir) {
			return fmt.Errorf("missing code or policy dir")
		}
		scanner, er = policy.NewScannerFromLocalDir(c.TemplateDir, c.PolicyDir)
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

	err := scanner.Run()
	if err != nil {
		return err
	}

	return nil
}

func (c *ScanCmd) Parse(filePath string) error {
	cmdString := utils.SprintTemplate("terrascan scan --parse-only -d . -o json > {{.TerrascanResultFile}}", map[string]interface{}{
		"TFScanJsonFilePath": filepath.Join("./", runner.TerrascanJsonFile),
	})
	result, err := RunCmd(cmdString)
	if err != nil {
		return err
	}
	fmt.Printf("parse result: %s\n", result)
	return nil
}

func (c *ScanCmd) Scan(filePath string, regoDir string) error {
	cmdString := utils.SprintTemplate("terrascan scan -d . -o json > {{.TerrascanResultFile}}", map[string]interface{}{
		"TFScanJsonFilePath": filepath.Join("./", runner.TerrascanResultFile),
	})
	result, err := RunCmd(cmdString)
	if err != nil {
		return err
	}
	fmt.Printf("scan result: %s\n", result)
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
