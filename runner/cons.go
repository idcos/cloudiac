package runner

import "time"

var (
	AnsibleEnv = map[string]string{
		"ANSIBLE_HOST_KEY_CHECKING":"False",
		"ANSIBLE_TF_DIR":"..",
		"ANSIBLE_NOCOWS":"1",
	}
)

// 以下常量定义的是 runner 启动任务后容器内部的路径，不受配置文件响应

const ContainerWorkingDir = "/workspace"
const ContainerIaCDir = "/iac"
const ContainerProviderPath = "/providers"

const TaskLogName = "runner.log"
const TaskScriptName = "run.sh"
const BackendConfigName = "backend.tf"

const FollowLogDelay = time.Second // follow 文件时读到 EOF 后进行下次读取的等待时长

