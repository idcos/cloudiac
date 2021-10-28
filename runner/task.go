// Copyright 2021 CloudJ Company Limited. All rights reserved.

package runner

import (
	"bytes"
	"cloudiac/common"
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Task struct {
	req       RunTaskReq
	logger    logs.Logger
	config    configs.RunnerConfig
	workspace string
}

func NewTask(req RunTaskReq, logger logs.Logger) *Task {
	return &Task{
		req:    req,
		logger: logger,
		config: configs.Get().Runner,
	}
}

func (t *Task) Run() (cid string, err error) {
	if t.req.ContainerId == "" {
		cid, err = t.start()
		if err != nil {
			return cid, err
		}
		t.req.ContainerId = cid
	}
	return t.req.ContainerId, t.runStep()
}

func (t *Task) start() (cid string, err error) {
	for _, vars := range []map[string]string{
		t.req.Env.EnvironmentVars, t.req.Env.TerraformVars, t.req.Env.AnsibleVars} {
		if err = t.decryptVariables(vars); err != nil {
			return "", errors.Wrap(err, "decrypt variables")
		}
	}

	if t.req.PrivateKey != "" {
		t.req.PrivateKey, err = utils.DecryptSecretVar(t.req.PrivateKey)
		if err != nil {
			return "", errors.Wrap(err, "decrypt private key")
		}
	}

	t.workspace, err = t.initWorkspace()
	if err != nil {
		return "", errors.Wrap(err, "initial workspace")
	}

	conf := configs.Get().Runner
	cmd := Executor{
		Image:       conf.DefaultImage,
		Timeout:     t.req.Timeout,
		Workdir:     ContainerWorkspace,
		HostWorkdir: t.workspace,
	}

	if t.req.DockerImage != "" {
		cmd.Image = t.req.DockerImage
	}

	tfPluginCacheDir := ""
	for k, v := range t.req.Env.EnvironmentVars {
		if k == "TF_PLUGIN_CACHE_DIR" {
			tfPluginCacheDir = v
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	if tfPluginCacheDir == "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%s", ContainerPluginCachePath))
	}

	// 变量名冲突时，系统环境变量覆盖用户定义的环境变量
	for k, v := range t.req.SysEnvironments {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	for k, v := range t.req.Env.TerraformVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TF_VAR_%s=%s", k, v))
	}
	if t.req.Env.TfVersion == "" {
		t.req.Env.TfVersion = consts.DefaultTerraformVersion
	}
	cmd.TerraformVersion = t.req.Env.TfVersion
	cmd.Env = append(cmd.Env, fmt.Sprintf("TFENV_TERRAFORM_VERSION=%s", cmd.TerraformVersion))

	// 容器启动后执行 /bin/sh，以保持运行
	cmd.Commands = []string{"/bin/sh"}

	stepDir := GetTaskDir(t.req.Env.Id, t.req.TaskId, t.req.Step)
	containerInfoFile := filepath.Join(stepDir, TaskContainerInfoFileName)
	// 启动容器前先删除可能存在的 containerInfoFile，以支持步骤重试，
	// 否则 containerInfoFile 文件存在 CommittedTask.Wait() 会直接返回
	if err = os.Remove(containerInfoFile); err != nil && !os.IsNotExist(err) {
		return "", errors.Wrap(err, "remove containerInfoFile")
	}

	t.logger.Infof("start task step, workdir: %s", cmd.HostWorkdir)
	if cid, err = cmd.Start(); err != nil {
		return cid, err
	}

	return cid, nil
}

func (t *Task) generateCommand(cmd string) []string {
	cmds := []string{"/bin/sh"}
	if utils.IsTrueStr(t.req.Env.EnvironmentVars["CLOUDIAC_DEBUG"]) {
		cmds = append(cmds, "-x")
	}
	return append(cmds, "-c", cmd)
}

func (t *Task) runStep() (err error) {
	_, err = t.genStepScript()
	if err != nil {
		return errors.Wrap(err, "generate step script")
	}

	containerScriptPath := filepath.Join(t.stepDirName(t.req.Step), TaskScriptName)
	logPath := filepath.Join(t.stepDirName(t.req.Step), TaskLogName)
	command := fmt.Sprintf("%s >>%s 2>&1", containerScriptPath, logPath)

	if ok, err := (Executor{}).IsPaused(t.req.ContainerId); err != nil {
		return err
	} else if ok {
		logger.Debugf("container %s is paused", t.req.ContainerId)

		logger.Debugf("unpause container")
		if err := (Executor{}).Unpause(t.req.ContainerId); err != nil {
			return err
		}
		logger.Debugf("unpause container done")
	}

	execId, err := (&Executor{}).RunCommand(t.req.ContainerId, t.generateCommand(command))
	if err != nil {
		return err
	}

	// 后台协程监控到命令结束就会暂停容器，
	// 同时 task.Wait() 函数也会在任务结束后暂停容器，两边同时处理保证容器被暂停
	if t.req.PauseTask {
		go func() {
			_, err := (Executor{}).WaitCommand(context.Background(), execId)
			if err != nil {
				logger.Debugf("container %s: %v", t.req.ContainerId, err)
				return
			}

			logger.Debugf("pause container %s", t.req.ContainerId)
			if err := (&Executor{}).Pause(t.req.ContainerId); err != nil {
				logger.Debugf("container %s: %v", t.req.ContainerId, err)
			}
		}()
	}

	infoJson := utils.MustJSON(StartedTask{
		EnvId:         t.req.Env.Id,
		TaskId:        t.req.TaskId,
		Step:          t.req.Step,
		ContainerId:   t.req.ContainerId,
		PauseOnFinish: t.req.PauseTask,
		ExecId:        execId,
	})

	stepInfoFile := filepath.Join(
		GetTaskDir(t.req.Env.Id, t.req.TaskId, t.req.Step),
		TaskInfoFileName,
	)
	if err := os.WriteFile(stepInfoFile, infoJson, 0644); err != nil {
		err = errors.Wrap(err, "write step info")
		return err
	}
	return nil
}

func (t *Task) decryptVariables(vars map[string]string) error {
	var err error
	for k, v := range vars {
		vars[k], err = utils.DecryptSecretVar(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Task) initWorkspace() (workspace string, err error) {
	if strings.HasPrefix(t.req.Env.Workdir, "..") {
		// 不允许访问上层目录
		return "", fmt.Errorf("invalid workdir '%s'", t.req.Env.Workdir)
	}

	workspace = GetTaskWorkspace(t.req.Env.Id, t.req.TaskId)
	if t.req.Step != 0 {
		return workspace, nil
	}

	if err = os.MkdirAll(workspace, 0755); err != nil {
		return workspace, err
	}

	privateKeyPath := filepath.Join(workspace, "ssh_key")
	keyContent := fmt.Sprintf("%s\n", strings.TrimSpace(t.req.PrivateKey))
	if err = os.WriteFile(privateKeyPath, []byte(keyContent), 0600); err != nil {
		return workspace, err
	}

	if err = t.genIacTfFile(workspace); err != nil {
		return workspace, errors.Wrap(err, "generate tf file")
	}
	if err = t.genPlayVarsFile(workspace); err != nil {
		return workspace, errors.Wrap(err, "generate play vars file")
	}

	return workspace, nil
}

var iacTerraformTpl = template.Must(template.New("").Parse(` terraform {
  backend "{{.State.Backend}}" {
    address = "{{.State.Address}}"
    scheme  = "{{.State.Scheme}}"
    path    = "{{.State.Path}}"
    lock    = true
    gzip    = false
  }
}

locals {
	cloudiac_ssh_user    = "root"
	cloudiac_private_key = "{{.PrivateKeyPath}}"
}
`))

func execTpl2File(tpl *template.Template, data interface{}, savePath string) error {
	fp, err := os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer fp.Close()
	return tpl.Execute(fp, data)
}

func (t *Task) genIacTfFile(workspace string) error {
	if t.req.StateStore.Address == "" {
		if os.Getenv("IAC_WORKER_CONSUL") != "" {
			t.req.StateStore.Address = os.Getenv("IAC_WORKER_CONSUL")
		} else {
			t.req.StateStore.Address = configs.Get().Consul.Address
		}
	}
	ctx := map[string]interface{}{
		"Workspace":      workspace,
		"PrivateKeyPath": t.up2Workspace("ssh_key"),
		"State":          t.req.StateStore,
	}
	if err := execTpl2File(iacTerraformTpl, ctx, filepath.Join(workspace, CloudIacTfFile)); err != nil {
		return err
	}
	return nil
}

var iacPlayVarsTpl = template.Must(template.New("").Parse(`
{{- range $k,$v := .Env.AnsibleVars -}}
{{$k}} = "{{$v}}"
{{- end -}}
`))

func (t *Task) genPlayVarsFile(workspace string) error {
	fp, err := os.OpenFile(filepath.Join(workspace, CloudIacPlayVars), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	return yaml.NewEncoder(fp).Encode(t.req.Env.AnsibleVars)
}

func (t *Task) genPolicyFiles(workspace string) error {
	if len(t.req.Policies) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Join(workspace, PoliciesDir), 0755); err != nil {
		return err
	}
	for _, policy := range t.req.Policies {
		if err := os.MkdirAll(filepath.Join(workspace, PoliciesDir, policy.PolicyId), 0755); err != nil {
			return err
		}
		js, _ := json.Marshal(policy.Meta)
		if err := os.WriteFile(filepath.Join(workspace, PoliciesDir, policy.PolicyId, "meta.json"), js, 0644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(workspace, PoliciesDir, policy.PolicyId, "policy.rego"), []byte(policy.Rego), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (t *Task) executeTpl(tpl *template.Template, data interface{}) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := tpl.Execute(buffer, data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (t *Task) stepDirName(step int) string {
	return GetTaskDirName(step)
}

func (t *Task) genStepScript() (string, error) {
	var (
		command string
		err     error
	)

	switch t.req.StepType {
	case common.TaskStepCheckout:
		command, err = t.stepCheckout()
	case common.TaskStepTfInit:
		command, err = t.stepInit()
	case common.TaskStepTfPlan:
		command, err = t.stepPlan()
	case common.TaskStepTfApply:
		command, err = t.stepApply()
	case common.TaskStepTfDestroy:
		command, err = t.stepDestroy()
	case common.TaskStepAnsiblePlay:
		command, err = t.stepPlay()
	case common.TaskStepCommand:
		command, err = t.stepCommand()
	case common.TaskStepCollect:
		command, err = t.collectCommand()
	case common.TaskStepScanInit:
		command, err = t.stepScanInit()
	case common.TaskStepRegoParse:
		command, err = t.stepTfParse()
	case common.TaskStepOpaScan:
		command, err = t.stepTfScan()
	default:
		return "", fmt.Errorf("unknown step type '%s'", t.req.StepType)
	}
	if err != nil {
		return "", err
	}

	stepDir := GetTaskDir(t.req.Env.Id, t.req.TaskId, t.req.Step)
	if err = os.MkdirAll(stepDir, 0755); err != nil {
		return "", err
	}

	scriptPath := filepath.Join(stepDir, TaskScriptName)
	var fp *os.File
	if fp, err = os.OpenFile(scriptPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755); err != nil {
		return "", err
	}
	defer fp.Close()

	if _, err = fp.WriteString(command); err != nil {
		return "", err
	}

	return scriptPath, nil
}

var checkoutCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
if [[ ! -e code ]]; then git clone '{{.Req.RepoAddress}}' code; fi && \
cd code && \
echo 'checkout {{.Req.RepoCommitId}}.' && \
git checkout -q '{{.Req.RepoCommitId}}' && \
cd '{{.Req.Env.Workdir}}'
`))

func (t *Task) stepCheckout() (command string, err error) {
	return t.executeTpl(checkoutCommandTpl, map[string]interface{}{
		"Req": t.req,
	})
}

var initCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
cd 'code/{{.Req.Env.Workdir}}' && \
ln -sf '{{.IacTfFile}}' . && \
tfenv install $TFENV_TERRAFORM_VERSION && \
tfenv use $TFENV_TERRAFORM_VERSION  && \
terraform init -input=false {{- range $arg := .Req.StepArgs }} {{$arg}}{{ end }}
`))

// 将 workspace 根目录下的文件名转为可以在环境的 code/workdir 下访问的相对路径
func (t *Task) up2Workspace(name string) string {
	ups := make([]string, 0)
	ups = append(ups, "..") // 代码仓库被 clone 到 code 目录，所以默认有一层目录包装
	for range filepath.SplitList(t.req.Env.Workdir) {
		ups = append(ups, "..")
	}
	return filepath.Join(append(ups, name)...)
}

func (t *Task) stepInit() (command string, err error) {
	return t.executeTpl(initCommandTpl, map[string]interface{}{
		"Req":             t.req,
		"PluginCachePath": ContainerPluginCachePath,
		"IacTfFile":       t.up2Workspace(CloudIacTfFile),
	})
}

var planCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
cd 'code/{{.Req.Env.Workdir}}' && \
terraform plan -input=false -out=_cloudiac.tfplan \
{{if .TfVars}}-var-file={{.TfVars}}{{end}} \
{{ range $arg := .Req.StepArgs }}{{$arg}} {{ end }}&& \
terraform show -no-color -json _cloudiac.tfplan >{{.TFPlanJsonFilePath}}
`))

func (t *Task) stepPlan() (command string, err error) {
	return t.executeTpl(planCommandTpl, map[string]interface{}{
		"Req":                t.req,
		"TfVars":             t.req.Env.TfVarsFile,
		"TFPlanJsonFilePath": t.up2Workspace(TFPlanJsonFile),
	})
}

// 当指定了 plan 文件时不需要也不能传 -var-file 参数
var applyCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
cd 'code/{{.Req.Env.Workdir}}' && \
terraform apply -input=false -auto-approve \
{{ range $arg := .Req.StepArgs}}{{$arg}} {{ end }}_cloudiac.tfplan
`))

func (t *Task) stepApply() (command string, err error) {
	return t.executeTpl(applyCommandTpl, map[string]interface{}{
		"Req": t.req,
	})
}

func (t *Task) stepDestroy() (command string, err error) {
	// destroy 任务通过会先执行 plan(传入 --destroy 参数)，然后再 apply plan 文件实现。
	// 这样可以保证 destroy 时执行的是用户审批时看到的 plan 内容
	return t.executeTpl(applyCommandTpl, map[string]interface{}{
		"Req": t.req,
	})
}

var playCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
export ANSIBLE_HOST_KEY_CHECKING="False"
export ANSIBLE_TF_DIR="."
export ANSIBLE_NOCOWS="1"

cd 'code/{{.Req.Env.Workdir}}' && ansible-playbook \
--inventory {{.AnsibleStateAnalysis}} \
--user "root" \
--private-key "{{.PrivateKeyPath}}" \
--extra @{{.IacPlayVars}} \
{{ if .Req.Env.PlayVarsFile -}}
--extra @{{.Req.Env.PlayVarsFile}} \
{{ end -}}
{{ range $arg := .Req.StepArgs }}{{$arg}} {{ end }} \
{{.Req.Env.Playbook}} 
`))

func (t *Task) stepPlay() (command string, err error) {
	return t.executeTpl(playCommandTpl, map[string]interface{}{
		"Req":                  t.req,
		"IacPlayVars":          t.up2Workspace(CloudIacPlayVars),
		"PrivateKeyPath":       t.up2Workspace("ssh_key"),
		"AnsibleStateAnalysis": filepath.Join(ContainerAssetsDir, AnsibleStateAnalysisName),
	})
}

var cmdCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
(test -d 'code/{{.Req.Env.Workdir}}' && cd 'code/{{.Req.Env.Workdir}}')
{{ range $index, $command := .Commands -}}
{{$command}}
{{ end -}}
`))

func (t *Task) stepCommand() (command string, err error) {
	commands := make([]string, 0)
	for _, c := range t.req.StepArgs {
		commands = append(commands, fmt.Sprintf("%v", c))
	}

	return t.executeTpl(cmdCommandTpl, map[string]interface{}{
		"Req":      t.req,
		"Commands": commands,
	})
}

// collect command 失败不影响任务状态
var collectCommandTpl = template.Must(template.New("").Parse(`# state collect command
cd 'code/{{.Req.Env.Workdir}}' && \
terraform show -no-color -json >{{.TFStateJsonFilePath}} && \
terraform providers schema -json > {{.TFProviderSchema}}
`))

func (t *Task) collectCommand() (string, error) {
	return t.executeTpl(collectCommandTpl, map[string]interface{}{
		"Req":                 t.req,
		"TFStateJsonFilePath": t.up2Workspace(TFStateJsonFile),
		"TFProviderSchema":    t.up2Workspace(TFProviderSchema),
	})
}

var parseCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
cd 'code/{{.Req.Env.Workdir}}' && \
mkdir -p {{.PoliciesDir}} && \
mkdir -p ~/.terrascan/pkg/policies/opa/rego/aws && \
terrascan scan --config-only -l debug -o json > {{.TFScanJsonFilePath}}
`))

func (t *Task) stepTfParse() (command string, err error) {
	return t.executeTpl(parseCommandTpl, map[string]interface{}{
		"Req":                 t.req,
		"IacPlayVars":         t.up2Workspace(CloudIacPlayVars),
		"TFScanJsonFilePath":  t.up2Workspace(TerrascanJsonFile),
		"PoliciesDir":         t.up2Workspace(PoliciesDir),
		"TerrascanResultFile": t.up2Workspace(TerrascanResultFile),
	})
}

var scanCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
cd 'code/{{.Req.Env.Workdir}}' && \
mkdir -p {{.PoliciesDir}} && \
mkdir -p ~/.terrascan/pkg/policies/opa/rego/aws && \
echo scanning policies && \
terrascan scan -p {{.PoliciesDir}} --show-passed --iac-type terraform -l debug -o json > {{.TerrascanResultFile}}
`))

func (t *Task) stepTfScan() (command string, err error) {
	if err = t.genPolicyFiles(t.workspace); err != nil {
		return "", errors.Wrap(err, "generate policy files")
	}
	return t.executeTpl(scanCommandTpl, map[string]interface{}{
		"Req":                 t.req,
		"IacPlayVars":         t.up2Workspace(CloudIacPlayVars),
		"PoliciesDir":         t.up2Workspace(PoliciesDir),
		"TerrascanResultFile": t.up2Workspace(TerrascanResultFile),
	})
}

var scanInitCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
if [[ ! -e code ]]; then git clone '{{.Req.RepoAddress}}' code; fi && \
cd code && \
echo 'checkout {{.Req.RepoCommitId}}.' && \
git checkout -q '{{.Req.RepoCommitId}}' && \
cd '{{.Req.Env.Workdir}}'
`))

func (t *Task) stepScanInit() (command string, err error) {
	return t.executeTpl(scanInitCommandTpl, map[string]interface{}{
		"Req":             t.req,
		"PluginCachePath": ContainerPluginCachePath,
		"IacTfFile":       t.up2Workspace(CloudIacTfFile),
	})
}
