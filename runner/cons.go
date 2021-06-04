package runner

import "time"

var (
	AnsibleEnv = map[string]string{
		"ANSIBLE_HOST_KEY_CHECKING":"False",
		"ANSIBLE_TF_DIR":".",
		"ANSIBLE_NOCOWS":"1",
	}
)

// 以下常量定义的是 runner 启动任务后容器内部的路径，不受配置文件响应
const (
	AnsibleStateAnalysis = "/usr/yunji/cloudiac/terraform.py"
	ContainerWorkingDir = "/workspace"
	ContainerIaCDir = "/iac"
	ContainerProviderPath = "/providers"
	TaskLogName = "runner.log"
	TaskScriptName = "run.sh"
	BackendConfigName = "backend.tf"
	FollowLogDelay = time.Second // follow 文件时读到 EOF 后进行下次读取的等待时长
)


