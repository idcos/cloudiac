package runner

import "time"

var (
	AnsibleEnv = map[string]string{
		"ANSIBLE_HOST_KEY_CHECKING": "False",
		"ANSIBLE_TF_DIR":            ".",
		"ANSIBLE_NOCOWS":            "1",
	}
)

/*
provider 的加载逻辑:
1. 检查 cache 目录是否存在目标 provider，存在则直接使用，否则
2. 检查本地 plugins 目录(包含多个目录，具体参考下方文档)下是否存在，存在则拷贝一份(或创建软链接)到 cache 目录，否则
3. 从网络下载文件，并保存到 cache 目录

最后将 cache 目录的文件链接到当前目录的 .terraform/providers

参考文档:
- https://www.terraform.io/docs/cli/config/config-file.html#implied-local-mirror-directories
- https://www.terraform.io/docs/cli/config/config-file.html#provider-plugin-cache
*/

/////
// 以下常量定义的是 runner 启动任务后容器内部的路径，不受配置文件响应
const (
	AnsibleStateAnalysis     = "/usr/yunji/cloudiac/terraform.py"
	ContainerWorkspace       = "/workspace"
	ContainerIaCDir          = "/cloud_iac" // 任务相关文件(run.sh/log) 挂载目录
	ContainerProviderPath    = "/providers"
	ContainerPluginCachePath = "/terraform/cache/plugins" // terraform plugins(providers) 缓存目录
)

const (
	TaskLogName       = "runner.log"
	TaskScriptName    = "run.sh"
	BackendConfigName = "backend.tf"
	FollowLogDelay    = time.Second // follow 文件时读到 EOF 后进行下次读取的等待时长
)
