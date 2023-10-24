// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package runner

import (
	"bytes"
	"cloudiac/common"
	"cloudiac/configs"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Task struct {
	req    RunTaskReq
	logger logs.Logger
	// config    configs.RunnerConfig
	workspace string
}

func NewTask(req RunTaskReq, logger logs.Logger) *Task {
	return &Task{
		req:    req,
		logger: logger,
		// config: configs.Get().Runner,
	}
}

func CleanTaskWorkDirCode(envId, taskId string) error {
	logger.Debugf("CleanTaskWorkDirCode params: envId=%s, taskId=%s", envId, taskId)
	workspace := GetTaskWorkspace(envId, taskId)
	if workspace == "" {
		return nil
	}

	err := os.RemoveAll(filepath.Join(workspace, "code"))
	if err != nil {
		logger.Warnf("CleanTaskWorkDirCode error: %v\n", err)
	}

	return err
}

func (t *Task) Run() (cid string, err error) {
	if t.req.ContainerId == "" {
		cid, err = t.start()
		if err != nil {
			return cid, err
		}
		t.req.ContainerId = cid
	} else {
		// 初始化 workspace 路径名称
		t.workspace, err = t.initWorkspace()
		if err != nil {
			return "", errors.Wrap(err, "initial workspace")
		}
	}

	return t.req.ContainerId, t.runStep()
}

func (t *Task) start() (cid string, err error) {
	if t.req.PrivateKey != "" {
		t.req.PrivateKey, err = utils.DecryptSecretVar(t.req.PrivateKey)
		if err != nil {
			return "", errors.Wrap(err, "decrypt private key")
		}
	}
	for _, vars := range []map[string]string{
		t.req.Env.EnvironmentVars, t.req.Env.TerraformVars, t.req.Env.AnsibleVars} {
		if err := t.decryptVariables(vars); err != nil {
			return "", errors.Wrap(err, "decrypt variables")
		}
	}

	t.workspace, err = t.initWorkspace()
	if err != nil {
		return "", errors.Wrap(err, "initial workspace")
	}

	conf := configs.Get().Runner
	cmd := Executor{
		Image:       conf.DefaultImage,
		Name:        t.req.TaskId,
		Timeout:     t.req.Timeout,
		Workdir:     ContainerWorkspace,
		HostWorkdir: t.workspace,
		PluginCache: GetEnvPluginCache(t.req.Env.Id),
	}

	if t.req.DockerImage != "" {
		cmd.Image = t.req.DockerImage
	}

	reserveContainer := conf.ReserveContainer
	if v, ok := t.req.Env.EnvironmentVars["CLOUDIAC_RESERVER_CONTAINER"]; ok {
		// 需要明确判断是否为 true 或者 false，其他情况使用配置文件中的值
		if utils.IsTrueStr(v) {
			reserveContainer = true
		} else if utils.IsFalseStr(v) {
			reserveContainer = false
		}
	}
	cmd.AutoRemove = !reserveContainer

	if err := t.buildVarsAndCmdEnv(&cmd); err != nil {
		return "", err
	}

	// 容器启动后执行 /bin/bash 以保持运行，然后通过 exec 在容器中执行步骤命令
	cmd.Commands = []string{"/bin/bash"}

	stepDir := GetTaskDir(t.req.Env.Id, t.req.TaskId, t.req.Step)
	containerInfoFile := filepath.Join(stepDir, TaskContainerInfoFileName)
	// 启动容器前先删除可能存在的 containerInfoFile，以支持步骤重试，
	// 否则 containerInfoFile 文件存在 CommittedTask.Wait() 会直接返回
	if err = os.Remove(containerInfoFile); err != nil && !os.IsNotExist(err) {
		return "", errors.Wrap(err, "remove containerInfoFile")
	}

	t.logger.Infof("start task step, %s", stepDir)
	if cid, err = cmd.Start(); err != nil {
		return cid, err
	}

	return cid, nil
}

func (t *Task) buildVarsAndCmdEnv(cmd *Executor) error {
	// 设置默认的 LC_ALL，解决 ansible playbook 中输出中文乱码问题
	cmd.Env = append(cmd.Env, "LC_ALL=en_US.UTF-8")

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

	if t.req.Env.TfVersion == "" {
		t.req.Env.TfVersion = configs.Get().GetDefaultTerraformVersion()
	}
	cmd.TerraformVersion = t.req.Env.TfVersion
	cmd.Env = append(cmd.Env, fmt.Sprintf("TFENV_TERRAFORM_VERSION=%s", cmd.TerraformVersion))
	return nil
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

	var command string
	if utils.StrInArray(t.req.StepType, common.TaskStepCheckout, common.TaskStepScanInit) {
		// 移除日志中可能出现的 token 信息
		command = fmt.Sprintf("set -o pipefail\n%s 2>&1 >>%s", containerScriptPath, logPath)
	} else {
		command = fmt.Sprintf("%s >>%s 2>&1", containerScriptPath, logPath)
	}

	if t.req.Step >= 0 { // step < 0 表示是隐含步骤，不需要判断任务是否已中止
		if info, err := ReadTaskControlInfo(t.req.Env.Id, t.req.TaskId); err != nil {
			return err
		} else if info.Aborted() {
			return ErrTaskAborted
		}
	}

	if err := (Executor{}).UnpauseIf(t.req.ContainerId); err != nil {
		return err
	}

	execId, err := (&Executor{}).RunCommand(t.req.ContainerId, t.generateCommand(command))
	if err != nil {
		return err
	}

	now := time.Now()
	infoJson := utils.MustJSON(StepInfo{
		EnvId:         t.req.Env.Id,
		TaskId:        t.req.TaskId,
		Step:          t.req.Step,
		Workdir:       t.req.Env.Workdir,
		StatePath:     t.req.StateStore.Path,
		ContainerId:   t.req.ContainerId,
		PauseOnFinish: t.req.PauseTask,
		ExecId:        execId,
		StartedAt:     &now,
		Timeout:       t.req.Timeout,
	})

	stepInfoFile := filepath.Join(
		GetTaskDir(t.req.Env.Id, t.req.TaskId, t.req.Step),
		TaskStepInfoFileName,
	)
	latestStepInfoFile := filepath.Join(
		GetTaskWorkspace(t.req.Env.Id, t.req.TaskId),
		TaskStepInfoFileName,
	)

	if err := os.WriteFile(latestStepInfoFile, infoJson, 0644); err != nil { //nolint:gosec
		err = errors.Wrap(err, "write latest step info")
		return err
	}

	if err := os.WriteFile(stepInfoFile, infoJson, 0644); err != nil { //nolint:gosec
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

	if err = t.genEnvironmentFile(workspace); err != nil {
		return workspace, errors.Wrap(err, "generate environment file")
	}
	if err = t.genIacTfFile(workspace); err != nil {
		return workspace, errors.Wrap(err, "generate cloudiac tf file")
	}
	if err = t.genTfvarsJsonFile(workspace); err != nil {
		return workspace, errors.Wrap(err, "generate tfvars json file")
	}
	if err = t.genPlayVarsFile(workspace); err != nil {
		return workspace, errors.Wrap(err, "generate play vars file")
	}
	if err = t.genTerraformrcFile(workspace); err != nil {
		return workspace, errors.Wrap(err, "generate terraformrc file")
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
{{if .State.ConsulAcl}}access_token = "{{.State.ConsulToken}}"{{end}}
    {{if .State.ConsulTls}}
    ca_file = "{{.State.CaPath}}"
    cert_file = "{{.State.CapemPath}}"
    key_file = "{{.State.CakeyPath}}"
	{{end}}
  }
}

locals {
	cloudiac_ssh_user    = "root"
	cloudiac_private_key = "{{.PrivateKeyPath}}"
}
`))

func execTpl2File(tpl *template.Template, data interface{}, savePath string) error {
	fp, err := os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644) //nolint:gosec
	if err != nil {
		return err
	}
	defer fp.Close()

	return tpl.Execute(fp, data)
}

// 记录执行任务时使用的系统环境变量
func (t *Task) genEnvironmentFile(workspace string) error {
	path := filepath.Join(workspace, EnvironmentFile)
	b := utils.MustJSONIndent(t.req.SysEnvironments, "  ")
	return os.WriteFile(path, b, 0644) //nolint:gosec
}

func (t *Task) genIacTfFile(workspace string) error {
	if t.req.StateStore.Address == "" {
		if os.Getenv("IAC_WORKER_CONSUL") != "" {
			t.req.StateStore.Address = os.Getenv("IAC_WORKER_CONSUL")
			//t.req.StateStore.Backend = "consul"
			//t.req.StateStore.Scheme = "http"
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

func (t *Task) genPlayVarsFile(workspace string) error {
	fp, err := os.OpenFile(filepath.Join(workspace, CloudIacPlayVars), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		_ = fp.Close()
	}()

	var ansibleVars = t.req.Env.AnsibleVars
	for key, value := range t.req.SysEnvironments {
		if key != "" && strings.HasPrefix(key, "CLOUDIAC_") {
			ansibleVars[strings.ToLower(key)] = value
		}
	}
	return yaml.NewEncoder(fp).Encode(ansibleVars)
}

func (t *Task) genTfvarsJsonFile(workspace string) error {
	fp, err := os.OpenFile(
		filepath.Join(workspace, CloudIacTfvarsJson),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0644) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		_ = fp.Close()
	}()

	vars := make(map[string]interface{})
	for k, v := range t.req.Env.TerraformVars {
		{ // 尝试将值做为 json 解析
			// 这里只需要处理 map、list、null 这三类特殊变量，
			// 其他变量类型都可以以字符串传入，terraform 可以正常处理
			tv := strings.TrimSpace(v)
			if strings.HasPrefix(tv, "{") {
				mv := make(map[string]interface{})
				if err := json.Unmarshal([]byte(tv), &mv); err == nil {
					vars[k] = mv
					continue
				}
			} else if strings.HasPrefix(tv, "[") {
				lv := make([]interface{}, 0)
				if err := json.Unmarshal([]byte(tv), &lv); err == nil {
					vars[k] = lv
					continue
				}
			} else if tv == "null" {
				vars[k] = nil
				continue
			}
		}
		vars[k] = v
	}

	enc := json.NewEncoder(fp)
	enc.SetIndent("", "  ")
	return enc.Encode(vars)
}

/*
    network mirror 段添加了 exclude = ["registry.terraform.io/idcos/*"]，
	因为 idcos 这个命名空间是我们之前特殊处理的，在 registry.terraform.io 上不存在（即使存在也不属于我们管理），
	所以当启用 network mirror 的时候也需要排除掉。
*/
var terraformrcTpl = template.Must(template.New("").Parse(`provider_installation {
  filesystem_mirror {
    path = "/cloudiac/terraform/plugins"
  }

  {{ if .NetworkMirrorUrl }}
  network_mirror {
    url = "{{.NetworkMirrorUrl}}"
    include = ["registry.terraform.io/*/*"]
    exclude = ["registry.terraform.io/idcos/*"]
  }
  {{ end }}

  direct {
    exclude = ["{{ .DirectExclude }}"]
  }
}`))

func (t *Task) genTerraformrcFile(workspace string) error {
	path := filepath.Join(workspace, TerraformrcFileName)

	// 默认情况下我们只针对 idcos 命名空间下的 provider 禁用 terraform 官方 registry
	// （如果不主动禁用，terraform cli 的默认行为总是会查询官方 registry 获取 provider 版本列表）
	directExclude := "registry.terraform.io/idcos/*"
	offline := configs.Get().Runner.OfflineMode
	if offline || t.req.NetworkMirror != "" {
		// 如果开启了 offline 或者 network mirror 则全局禁用 terraform 默认 registry
		directExclude = "registry.terraform.io/*/*"
	}

	return execTpl2File(terraformrcTpl, map[string]interface{}{
		"NetworkMirrorUrl": t.req.NetworkMirror,
		"DirectExclude":    directExclude,
	}, path)
}

func (t *Task) genPolicyFiles(workspace string) error {
	if len(t.req.Policies) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Join(workspace, PoliciesDir), 0755); err != nil { //nolint:gosec
		return err
	}
	for _, policy := range t.req.Policies {
		if err := os.MkdirAll(filepath.Join(workspace, PoliciesDir, policy.PolicyId), 0755); err != nil { //nolint:gosec
			return err
		}
		js, _ := json.Marshal(policy.Meta)
		if err := os.WriteFile(filepath.Join(workspace, PoliciesDir, policy.PolicyId, policy.Meta.Name+".json"), js, 0644); err != nil { //nolint:gosec
			return err
		}
		if err := os.WriteFile(filepath.Join(workspace, PoliciesDir, policy.PolicyId, policy.Meta.Name+".rego"), []byte(policy.Rego), 0644); err != nil { //nolint:gosec
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

//nolint:cyclop
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
	case common.TaskStepOpaScan:
		// 兼容 0.3 版本 pipeline
		// 为了保证 step envScan 步骤的正确运行，会自动插入 plan 步骤
		// 该行为会导致执行两次 plan，导致执行速度变慢，作为一个兼容性的已知问题
		var planCommand, scanCommand string
		if planCommand, err = t.stepPlan(); err == nil {
			if scanCommand, err = t.stepEnvScan(); err == nil {
				// 多个流程间执行需要退回到工作目录
				command = fmt.Sprintf("%s\ncd %s\n%s", planCommand, ContainerWorkspace, scanCommand)
			}
		}
	case common.TaskStepEnvParse:
		command, err = t.stepEnvParse()
	case common.TaskStepEnvScan:
		command, err = t.stepEnvScan()
	case common.TaskStepTplParse:
		command, err = t.stepTplParse()
	case common.TaskStepTplScan:
		command, err = t.stepTplScan()
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
set -o pipefail
# before checkout
cd '{{.ContainerWorkspace}}'
{{- if .Before }}
{{.Before}}
{{- end}}

if [ $? -ne 0 ]; then
	exit $?
fi

# goback to container workspace
cd {{.ContainerWorkspace}}
# clone code
git clone '{{.Req.RepoAddress}}' code 2>&1 | sed -re 's#(://[^:]+:)[^@]+#\1******#'
clone_result=$?

mkdir -p code && cd code
# clone success
if [ $clone_result -eq 0 ]; then
	echo 'checkout {{.Req.RepoCommitId}}.' && \
	git checkout -q '{{.Req.RepoCommitId}}'
fi

# create workdir in spite of clone was failed or not
mkdir -p '{{.Req.Env.Workdir}}' && cd '{{.Req.Env.Workdir}}'

ln -sf '{{.IacTfFile}}' && ln -sf '{{.TerraformRcFile}}' ~/.terraformrc

# after checkout
{{- if .After }}
{{.After}}
{{- end}}

if [ $? -ne 0 ]; then
	exit $?
fi

exit $clone_result
`))

func (t *Task) stepCheckout() (command string, err error) {
	beforeCmds, afterCmds := getBeforeAfterCmds(t.req.StepBeforeCmds, t.req.StepAfterCmds)
	return t.executeTpl(checkoutCommandTpl, map[string]interface{}{
		"Req":                t.req,
		"IacTfFile":          t.up2Workspace(CloudIacTfFile),
		"TerraformRcFile":    filepath.Join(ContainerWorkspace, TerraformrcFileName),
		"Before":             beforeCmds,
		"After":              afterCmds,
		"ContainerWorkspace": ContainerWorkspace,
	})
}

var initCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
cd '{{.ContainerWorkspace}}/code/{{.Req.Env.Workdir}}' && \
{{if .Before}}{{.Before}} && \{{- end}}
cd '{{.ContainerWorkspace}}/code/{{.Req.Env.Workdir}}' && \
tfenv install $TFENV_TERRAFORM_VERSION && \
tfenv use $TFENV_TERRAFORM_VERSION  && \
terraform init -input=false {{- range $arg := .Req.StepArgs }} {{$arg}}{{ end }} {{- if .After}} && \
{{.After}}{{- end}}
`))

// 将 workspace 根目录下的文件名转为可以在环境的 code/workdir 下访问的相对路径
func (t *Task) up2Workspace(name string) string {
	ups := make([]string, 0)
	ups = append(ups, "..") // 代码仓库被 clone 到 code 目录，所以默认有一层目录包装
	for _, v := range strings.Split(t.req.Env.Workdir, string(os.PathSeparator)) {
		if v != "" {
			ups = append(ups, "..")
		}
	}
	return filepath.Join(append(ups, name)...)
}

func (t *Task) stepInit() (command string, err error) {
	beforeCmds, afterCmds := getBeforeAfterCmds(t.req.StepBeforeCmds, t.req.StepAfterCmds)
	return t.executeTpl(initCommandTpl, map[string]interface{}{
		"Req":                t.req,
		"PluginCachePath":    ContainerPluginCachePath,
		"Before":             beforeCmds,
		"After":              afterCmds,
		"ContainerWorkspace": ContainerWorkspace,
	})
}

var planCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
cd '{{.ContainerWorkspace}}/code/{{.Req.Env.Workdir}}' && \
{{if .Before}}{{.Before}} && \{{- end}}
cd '{{.ContainerWorkspace}}/code/{{.Req.Env.Workdir}}' && \
terraform plan -detailed-exitcode -input=false -out=_cloudiac.tfplan \
{{if .TfVars}}-var-file={{.TfVars}} {{end}}-var-file={{.IacTfVars}} \
{{ range $arg := .Req.StepArgs }}{{$arg}} {{ end }}
status=$?
terraform show -no-color -json _cloudiac.tfplan >{{.TFPlanJsonFilePath}}
if [[ "$status" == "0" ]]; then
  echo "+--------+--------------------------------------------+"
  echo "| CHANGE |                    NAME                    |"
  echo "+--------+--------------------------------------------+"
  echo "| No changes.                                         |"
  echo "+-----------------------------------------------------+" 
elif [[ "$status" == "2" ]]; then 
  tf-summarize {{.TFPlanJsonFilePath}} 
else 
  exit $status
fi
{{- if .After}}{{.After}}{{- end}}
`))

func (t *Task) stepPlan() (command string, err error) {
	beforeCmds, afterCmds := getBeforeAfterCmds(t.req.StepBeforeCmds, t.req.StepAfterCmds)
	return t.executeTpl(planCommandTpl, map[string]interface{}{
		"Req":                t.req,
		"TfVars":             t.req.Env.TfVarsFile,
		"IacTfVars":          t.up2Workspace(CloudIacTfvarsJson),
		"TFPlanJsonFilePath": t.up2Workspace(TFPlanJsonFile),
		"Before":             beforeCmds,
		"After":              afterCmds,
		"ContainerWorkspace": ContainerWorkspace,
	})
}

// 当指定了 plan 文件时不需要也不能传 -var-file 参数
var applyCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
cd '{{.ContainerWorkspace}}/code/{{.Req.Env.Workdir}}' && \
{{if .Before}}{{.Before}} && \{{- end}}
cd '{{.ContainerWorkspace}}/code/{{.Req.Env.Workdir}}' && \
terraform apply -input=false -auto-approve \
{{ range $arg := .Req.StepArgs}}{{$arg}} {{ end }}_cloudiac.tfplan {{- if .After}} && \
{{.After}}{{- end}}

result=$?

# state collect command
cd '{{.ContainerWorkspace}}/code/{{.Req.Env.Workdir}}' && \
terraform show -no-color -json >{{.TFStateJsonFilePath}} && \
terraform providers schema -json > {{.TFProviderSchema}}
exit $result
`))

func (t *Task) stepApply() (command string, err error) {
	beforeCmds, afterCmds := getBeforeAfterCmds(t.req.StepBeforeCmds, t.req.StepAfterCmds)
	return t.executeTpl(applyCommandTpl, map[string]interface{}{
		"Req":                 t.req,
		"TFStateJsonFilePath": t.up2Workspace(TFStateJsonFile),
		"TFProviderSchema":    t.up2Workspace(TFProviderSchema),
		"Before":              beforeCmds,
		"After":               afterCmds,
		"ContainerWorkspace":  ContainerWorkspace,
	})
}

func (t *Task) stepDestroy() (command string, err error) {
	// destroy 任务通过会先执行 plan(传入 --destroy 参数)，然后再 apply plan 文件实现。
	// 这样可以保证 destroy 时执行的是用户审批时看到的 plan 内容
	beforeCmds, afterCmds := getBeforeAfterCmds(t.req.StepBeforeCmds, t.req.StepAfterCmds)
	return t.executeTpl(applyCommandTpl, map[string]interface{}{
		"Req":                 t.req,
		"TFStateJsonFilePath": t.up2Workspace(TFStateJsonFile),
		"TFProviderSchema":    t.up2Workspace(TFProviderSchema),
		"Before":              beforeCmds,
		"After":               afterCmds,
		"ContainerWorkspace":  ContainerWorkspace,
	})
}

// CLOUDIAC_WORKDIR 环境变量在 task_manager 中会自动设置
var playCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
export CLOUDIAC_ANSIBLE_INVENTORY={{.AnsibleStateAnalysis}}

{{if .Before}}cd "${CLOUDIAC_WORKDIR}" && {{.Before}} && \{{- end}}
cd "${CLOUDIAC_WORKDIR}" && \
if [[ -f "{{.Requirements}}" ]];then ansible-galaxy install -r "{{.Requirements}}"; fi && \
cloudiac-playbook \
{{ if .Req.Env.PlayVarsFile -}}--extra @{{.Req.Env.PlayVarsFile}}{{ end -}} \
{{ range $arg := .Req.StepArgs }}{{$arg}} {{ end }} \
{{.Req.Env.Playbook}} {{- if .After}} && \
{{.After}}{{- end}}
`))

func (t *Task) stepPlay() (command string, err error) {
	beforeCmds, afterCmds := getBeforeAfterCmds(t.req.StepBeforeCmds, t.req.StepAfterCmds)
	return t.executeTpl(playCommandTpl, map[string]interface{}{
		"Req":                  t.req,
		"Requirements":         filepath.Join(filepath.Dir(t.req.Env.Playbook), CloudIacAnsibleRequirements),
		"AnsibleStateAnalysis": filepath.Join(ContainerAssetsDir, AnsibleStateAnalysisName),
		"Before":               beforeCmds,
		"After":                afterCmds,
	})
}

var cmdCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
test -d 'code/{{.Req.Env.Workdir}}' && cd 'code/{{.Req.Env.Workdir}}'
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

var parseTplCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
cd 'code/{{.Req.Env.Workdir}}' && \
mkdir -p {{.PoliciesDir}} && \
mkdir -p ~/.terrascan/pkg/policies/opa/rego/aws && \
terrascan scan --config-only -l debug -o json --iac-type terraform > {{.ScanInputFile}}
`))

func (t *Task) stepTplParse() (command string, err error) {
	return t.executeTpl(parseTplCommandTpl, map[string]interface{}{
		"Req":           t.req,
		"ScanInputFile": t.up2Workspace(ScanInputFile),
		"PoliciesDir":   t.up2Workspace(PoliciesDir),
	})
}

var scanTplCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
cd 'code/{{.Req.Env.Workdir}}' && \
mkdir -p {{.PoliciesDir}} && \
mkdir -p ~/.terrascan/pkg/policies/opa/rego/aws && \
terrascan scan --config-only -o json --iac-type terraform > {{.ScanInputFile}} 2>/dev/null && \
/usr/yunji/cloudiac/iac-tool scan --internal -p {{.PoliciesDir}} -i {{.ScanInputFile}} -o {{.ScanResultFile}}
`))

func (t *Task) stepTplScan() (command string, err error) {
	if err = t.genPolicyFiles(t.workspace); err != nil {
		return "", errors.Wrap(err, "generate policy files")
	}
	return t.executeTpl(scanTplCommandTpl, map[string]interface{}{
		"Req":            t.req,
		"PoliciesDir":    t.up2Workspace(PoliciesDir),
		"ScanResultFile": t.up2Workspace(ScanResultFile),
		"ScanInputFile":  t.up2Workspace(ScanInputFile),
	})
}

var scanInitCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
if [[ ! -e code ]]; then git clone '{{.Req.RepoAddress}}' code || exit $?; fi && \
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

var envParseCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
cd 'code/{{.Req.Env.Workdir}}' && \
/usr/yunji/cloudiac/iac-tool scan --parse-plan --plan {{.TerraformPlanFile}} > {{.ScanInputFile}}
`))

func (t *Task) stepEnvParse() (command string, err error) {
	return t.executeTpl(envParseCommandTpl, map[string]interface{}{
		"Req":               t.req,
		"TerraformPlanFile": t.up2Workspace(TFPlanJsonFile),
		"ScanInputFile":     t.up2Workspace(ScanInputFile),
		"PoliciesDir":       t.up2Workspace(PoliciesDir),
	})
}

var envScanCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
#!/bin/sh
cd 'code/{{.Req.Env.Workdir}}' && \
mkdir -p {{.PoliciesDir}} && \
mkdir -p ~/.terrascan/pkg/policies/opa/rego/aws && \
terrascan scan --config-only -o json --iac-type terraform > {{.ScanInputMapFile}} 2>/dev/null && \
/usr/yunji/cloudiac/iac-tool scan --parse-plan --plan {{.TerraformPlanFile}} > {{.ScanInputFile}} && \
/usr/yunji/cloudiac/iac-tool scan --internal -p {{.PoliciesDir}} -i {{.ScanInputFile}} -m {{.ScanInputMapFile}} -o {{.ScanResultFile}}
`))

func (t *Task) stepEnvScan() (command string, err error) {
	if err = t.genPolicyFiles(t.workspace); err != nil {
		return "", errors.Wrap(err, "generate policy files")
	}
	return t.executeTpl(envScanCommandTpl, map[string]interface{}{
		"Req":               t.req,
		"TerraformPlanFile": t.up2Workspace(TFPlanJsonFile),
		"IacPlayVars":       t.up2Workspace(CloudIacPlayVars),
		"PoliciesDir":       t.up2Workspace(PoliciesDir),
		"ScanResultFile":    t.up2Workspace(ScanResultFile),
		"ScanInputFile":     t.up2Workspace(ScanInputFile),
		"ScanInputMapFile":  t.up2Workspace(ScanInputMapFile),
	})
}

func getBeforeAfterCmds(StepBeforeCmds, StepAfterCmds []string) (string, string) {
	beforeCmds := ""

	if len(StepBeforeCmds) > 0 {
		cmds := make([]string, 0)
		cmds = append(cmds, `echo "===== before commands ====="`)
		for _, cmd := range StepBeforeCmds {
			cmds = append(cmds, strings.TrimRight(cmd, "\n"))
		}
		cmds = append(cmds, `echo -e "===== end =====\n"`)
		beforeCmds = strings.Join(cmds, " && ")
	}

	afterCmds := ""
	if len(StepAfterCmds) > 0 {
		cmds := make([]string, 0)
		cmds = append(cmds, `echo -e "\n===== after commands ====="`)
		for _, cmd := range StepAfterCmds {
			cmds = append(cmds, strings.TrimRight(cmd, "\n"))
		}
		cmds = append(cmds, `echo "===== end ====="`)
		afterCmds = strings.Join(cmds, " && ")
	}

	return beforeCmds, afterCmds
}
