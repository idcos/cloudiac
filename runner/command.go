package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var initCommandTemplate = `set -e 
export CLOUD_IAC_TASK_DIR={{.TaskDir}}
export CLOUD_IAC_WORKSPACE={{.Workspace}}
export CLOUD_IAC_SSH_USER="root"
export CLOUD_IAC_PRIVATE_KEY={{.TaskDir}}/ssh_key

export TF_PLUGIN_CACHE_DIR={{.PluginsCachePath}}

git clone {{.Repo}} ${CLOUD_IAC_WORKSPACE} && \
cd "${CLOUD_IAC_WORKSPACE}" && git checkout {{.RepoCommit}} && \
ln -sv ${CLOUD_IAC_TASK_DIR}/{{.CloudIacTFName}}  ./ && \
terraform init && \`

const planCommandTemplate = `
terraform plan -input=false {{if .VarFile}}-var-file={{.VarFile}}{{end}}
`

const applyCommandTemplate = `
terraform apply -input=false -auto-approve {{if .VarFile}}-var-file={{.VarFile}}{{end}} && \
terraform state list >{{.ContainerStateListPath}} 2>&1 {{- if .AnsiblePlaybook}} && (
  export ANSIBLE_TF_DIR="${CLOUD_IAC_WORKSPACE}"
  cd {{.AnsibleWorkdir}} && ansible-playbook \
    --inventory {{.AnsibleStateAnalysis}} \
    --user "${CLOUD_IAC_SSH_USER}" \
    --private-key "${CLOUD_IAC_PRIVATE_KEY}" \
    {{.AnsiblePlaybook}}
)
{{- end}}
`

const destroyCommandTemplate = `
terraform destroy -input=false -auto-approve {{if .VarFile}}-var-file={{.VarFile}}{{end}} && \
terraform state list > {{.ContainerStateListPath}} 2>&1
`

const pullCommandTemplate = `
terraform state pull
`

var (
	initCommandTpl = template.Must(template.New("").Parse(initCommandTemplate))

	planCommandTpl    = template.Must(template.New("").Parse(planCommandTemplate))
	applyCommandTpl   = template.Must(template.New("").Parse(applyCommandTemplate))
	destroyCommandTpl = template.Must(template.New("").Parse(destroyCommandTemplate))
	pullCommandTpl    = template.Must(template.New("").Parse(pullCommandTemplate))

	commandTplMap = map[string]*template.Template{
		"plan":    planCommandTpl,
		"apply":   applyCommandTpl,
		"destroy": destroyCommandTpl,
		"pull":    pullCommandTpl,
	}
)

func GenScriptContent(context *ReqBody, saveTo string) error {
	saveFp, err := os.OpenFile(saveTo, os.O_CREATE|os.O_WRONLY, 0755)

	if err != nil {
		return err
	}
	defer saveFp.Close()

	isRunDebug := false
	// 允许为模板设置环境变量 IAC_DEBUG_TASK=1 来执行预设置的调试命令
	if isDebug, ok := context.Env["IAC_DEBUG_TASK"]; ok && (isDebug == "1" || strings.ToLower(isDebug) == "true") {
		isRunDebug = true
	} else if context.Mode == "debug" { // 或者直接传 mode=debug
		isRunDebug = true
	}

	if isRunDebug {
		// runner 启动时可以通过 IAC_DEBUG_COMMAND 环境变量自定义 debug 命令。
		// !!该变量不允许通过任务环境变量传入，避免任意命令执行
		debugCommand := os.Getenv("IAC_DEBUG_COMMAND")
		if debugCommand == "" {
			// 允许通过模板环境变量设置 count
			count := context.Env["IAC_DEBUG_RUN_COUNT"]
			if count == "" {
				count = "60"
			}
			debugCommand = fmt.Sprintf("for I in `seq 1 %s`;do date && hostname && uptime; sleep 1; done", count)
		}

		if _, err := fmt.Fprintf(saveFp, "\n%s\n", debugCommand); err != nil {
			return err
		}
		return nil
	}

	if err := initCommandTpl.Execute(saveFp, map[string]string{
		"Repo":             context.Repo,
		"RepoCommit":       context.RepoCommit,
		"Workspace":        ContainerWorkspace,
		"TaskDir":          ContainerTaskDir,
		"PluginsCachePath": ContainerPluginsCachePath,
		"CloudIacTFName":   CloudIacTFName,
	}); err != nil {
		return err
	}

	commandTpl, ok := commandTplMap[context.Mode]
	if !ok {
		return fmt.Errorf("unsupported mode '%s'", context.Mode)
	}

	var (
		containerStateListPath = filepath.Join(ContainerTaskDir, TerraformStateListName)
		ansibleWorkdir         = ""
		playbookName           = ""
	)
	if context.Playbook != "" {
		ansibleWorkdir = filepath.Dir(context.Playbook)
		playbookName = filepath.Base(context.Playbook)
	}
	if err := commandTpl.Execute(saveFp, map[string]string{
		"VarFile": context.Varfile,
		// 存储terraform state list输出内容弄的文件路径
		"ContainerStateListPath": containerStateListPath,
		"AnsibleWorkdir":         ansibleWorkdir,
		"AnsiblePlaybook":        playbookName,
		"AnsibleStateAnalysis":   filepath.Join(ContainerAssetsDir, AnsibleStateAnalysisName),
	}); err != nil {
		return err
	}

	if context.Extra != "" {
		if _, err := fmt.Fprintf(saveFp, "\n%s\n", context.Extra); err != nil {
			return err
		}
	}
	return nil
}
