package runner

import (
	"bytes"
	"cloudiac/common"
	"cloudiac/configs"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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
		return cid, errors.Wrap(err, "initial workspace")
	}

	if err = t.genStepScript(); err != nil {
		return cid, errors.Wrap(err, "generate step script")
	}

	conf := configs.Get().Runner
	cmd := Command{
		Image:       conf.DefaultImage,
		Env:         nil,
		Commands:    nil,
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

	cmd.Commands = []string{"sh", "-c", fmt.Sprintf("sh %s >>%s 2>&1",
		filepath.Join(t.stepDirName(t.req.Step), TaskStepScriptName),
		filepath.Join(t.stepDirName(t.req.Step), TaskStepLogName),
	)}

	t.logger.Infof("start task step, workdir: %s", cmd.HostWorkdir)
	if cid, err = cmd.Start(); err != nil {
		return cid, err
	}

	infoJson := utils.MustJSON(CommittedTaskStep{
		EnvId:       t.req.Env.Id,
		TaskId:      t.req.TaskId,
		Step:        t.req.Step,
		ContainerId: cid,
	})

	stepDir := GetTaskStepDir(t.req.Env.Id, t.req.TaskId, t.req.Step)
	if er := os.WriteFile(filepath.Join(stepDir, TaskStepInfoFileName), infoJson, 0644); er != nil {
		logger.Errorln(er)
	}
	return cid, err
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

	if ok, err := PathExists(workspace); err != nil {
		return workspace, err
	} else if ok && t.req.StepType == common.TaskStepInit {
		return workspace, fmt.Errorf("workspace '%s' is already exists", workspace)
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
	if err = t.genIacTfVarsFile(workspace); err != nil {
		return workspace, errors.Wrap(err, "generate tfvars file")
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
		t.req.StateStore.Address = configs.Get().Consul.Address
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

var iacTfVarsTpl = template.Must(template.New("").Parse(`
{{- range $k,$v := .Env.TerraformVars -}}
{{$k}} = "{{$v}}"
{{ end -}}
`))

func (t *Task) genIacTfVarsFile(workspace string) error {
	return execTpl2File(iacTfVarsTpl, t.req, filepath.Join(workspace, CloudIacTfVars))
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

func (t *Task) executeTpl(tpl *template.Template, data interface{}) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := tpl.Execute(buffer, data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (t *Task) stepDirName(step int) string {
	return GetTaskStepDirName(step)
}

func (t *Task) genStepScript() error {
	var (
		command string
		err     error
	)

	var collCmd string
	switch t.req.StepType {
	case common.TaskStepInit:
		command, err = t.stepInit()
	case common.TaskStepPlan:
		command, err = t.stepPlan()
	case common.TaskStepApply:
		command, err = t.stepApply()
		if err == nil {
			collCmd, err = t.collectCommand()
		}
	case common.TaskStepDestroy:
		command, err = t.stepDestroy()
		if err == nil {
			collCmd, err = t.collectCommand()
		}
	case common.TaskStepPlay:
		command, err = t.stepPlay()
	case common.TaskStepCommand:
		command, err = t.stepCommand()
	default:
		return fmt.Errorf("unknown step type '%s'", t.req.StepType)
	}
	if err != nil {
		return err
	}

	// 直接追加到末尾即可，即使步骤执行失败也执行信息采集命令
	if collCmd != "" {
		command = fmt.Sprintf("%s\n%s", command, collCmd)
	}

	stepDir := GetTaskStepDir(t.req.Env.Id, t.req.TaskId, t.req.Step)
	if err = os.MkdirAll(stepDir, 0755); err != nil {
		return err
	}

	scriptPath := filepath.Join(stepDir, TaskStepScriptName)
	var fp *os.File
	if fp, err = os.OpenFile(scriptPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString(command); err != nil {
		return err
	}

	return nil
}

var initCommandTpl = template.Must(template.New("").Parse(`#!/bin/sh
git clone '{{.Req.RepoAddress}}' code && \
cd 'code/{{.Req.Env.Workdir}}' && \
git checkout '{{.Req.RepoRevision}}' && \
ln -sf {{.IacTfFile}} . && \
ln -sf {{.IacTfVars}} . && \
terraform init -input=false {{- range $arg := .Req.StepArgs }} {{$arg}}{{ end }}
`))

// 将 workspace 根目录下的文件名转为可以在环境的 workdir 下访问的相对路径
func (t *Task) up2Workspace(name string) string {
	ups := make([]string, 0)
	ups = append(ups, "..")
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
		"IacTfVars":       t.up2Workspace(CloudIacTfVars),
	})
}

var planCommandTpl = template.Must(template.New("").Parse(`#/bin/sh
cd 'code/{{.Req.Env.Workdir}}' && \
terraform plan -input=false -out=_cloudiac.tfplan \
{{if .TfVars}}-var-file={{.TfVars}}{{end}} -var-file={{.IacTfVars}} \
{{ range $arg := .Req.StepArgs }}{{$arg}} {{ end }}&& \
terraform show -no-color -json _cloudiac.tfplan >{{.TFPlanJsonFilePath}}
`))

func (t *Task) stepPlan() (command string, err error) {
	return t.executeTpl(planCommandTpl, map[string]interface{}{
		"Req":                t.req,
		"TfVars":             t.req.Env.TfVarsFile,
		"IacTfVars":          CloudIacTfVars,
		"TFPlanJsonFilePath": t.up2Workspace(TFPlanJsonFile),
	})
}

// 当指定了 plan 文件时不需要也不能传 -var-file 参数
var applyCommandTpl = template.Must(template.New("").Parse(`#/bin/sh
cd 'code/{{.Req.Env.Workdir}}' && \
terraform apply -input=false -auto-approve \
{{ range $arg := .Req.StepArgs}}{{$arg}} {{ end }}_cloudiac.tfplan
`))

func (t *Task) stepApply() (command string, err error) {
	return t.executeTpl(applyCommandTpl, map[string]interface{}{
		"Req":       t.req,
		"TfVars":    t.req.Env.TfVarsFile,
		"IacTfVars": CloudIacTfVars,
	})
}

var destroyCommandTpl = template.Must(template.New("").Parse(`#/bin/sh
cd 'code/{{.Req.Env.Workdir}}' && \
terraform destroy -input=false -auto-approve \
{{if .TfVars}}-var-file={{.TfVars}}{{end}} -var-file={{.IacTfVars}} \
{{ range $arg := .Req.StepArgs}}{{$arg}} {{end}}
`))

func (t *Task) stepDestroy() (command string, err error) {
	return t.executeTpl(destroyCommandTpl, map[string]interface{}{
		"Req":       t.req,
		"TfVars":    t.req.Env.TfVarsFile,
		"IacTfVars": CloudIacTfVars,
	})
}

var playCommandTpl = template.Must(template.New("").Parse(`#/bin/sh
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

var cmdCommandTpl = template.Must(template.New("").Parse(`#/bin/sh
(test -d 'code/{{.Req.Env.Workdir}}' && cd 'code/{{.Req.Env.Workdir}}')
{{ range $index, $command := .Commands -}}
{{$command}} && \
{{ end -}}
sleep 0
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
terraform show -no-color -json >{{.TFStateJsonFilePath}} || sleep 0
`))

func (t *Task) collectCommand() (string, error) {
	return t.executeTpl(collectCommandTpl, map[string]interface{}{
		"TFStateJsonFilePath": t.up2Workspace(TFStateJsonFile),
	})
}
