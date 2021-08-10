package main

import (
	"bytes"
	"cloudiac/runner"
	"cloudiac/utils"
	"fmt"
	"os/exec"
	"path/filepath"
)

type PolicyCmd struct {
	TemplateDir string `long:"template-dir" description:"template directory" required:"false"`
	PolicyDir   string `long:"policy-dir" description:"policy directory" required:"false"`
	ParseOnly   bool   `long:"parse" description:"parse template" required:"false"`
}

func (*PolicyCmd) Usage() string {
	return "<the-terraform-file.tf> <the-policy-file.rego>"
}

func (c *PolicyCmd) Execute(args []string) error {
	if c.ParseOnly {
		if len(args) < 1 {
			return fmt.Errorf("missing template file")
		}
		return c.Parse(args[0])
	}

	if len(args) < 2 {
		return c.Scan(args[0], args[1])
	}

	return fmt.Errorf("not implement")
}

func (c *PolicyCmd) Parse(filePath string) error {
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

func (c *PolicyCmd) Scan(filePath string, regoDir string) error {
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
