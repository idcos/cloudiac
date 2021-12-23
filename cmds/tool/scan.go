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
// 4. 调试 rego 脚本
//    iac-tool scan --debug xxx.tf xxx.rego

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
		if c.CodeDir == "" {
			c.CodeDir = "code"
		}
		if c.PolicyDir == "" {
			c.PolicyDir = "policies"
		}
		if !utils.FileExist(c.CodeDir) || !utils.FileExist(c.PolicyDir) {
			return fmt.Errorf("missing code or policy dir")
		}
		scanner, er = policy.NewScannerFromLocalDir(c.CodeDir, c.PolicyDir)
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
		"JsonFile":       runner.TerrascanJsonFile,
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
	_, err := policy.RegoParse(regoFile, inputPath)
	fmt.Printf("Execute rego parse return err %+v", err)
	return err
}
