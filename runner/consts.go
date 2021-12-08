// Copyright 2021 CloudJ Company Limited. All rights reserved.

package runner

import "time"

/*
provider plugin 的查找逻辑:
1. 检查 cache 目录是否存在目标 provider，存在则直接使用，否则
2. 检查本地 plugins 目录(包含多个目录，具体参考下方文档)下是否存在，存在则拷贝一份(或创建软链接)到 cache 目录，否则
3. 从网络下载文件，并保存到 cache 目录

最后将 cache 目录的文件链接到当前目录的 .terraform/providers

参考文档:
- https://www.terraform.io/docs/cli/config/config-file.html#implied-local-mirror-directories
- https://www.terraform.io/docs/cli/config/config-file.html#provider-plugin-cache
*/

/////
// 以下定义的是 runner 启动任务后容器内部的路径，直接以常量配置即可
const (
	ContainerWorkspace = "/cloudiac/workspace"

	ContainerAssetsDir       = "/cloudiac/assets"                  // 挂载依赖资源，如 terraform.py 等(己打包到 worker 镜像)
	ContainerPluginPath      = "/cloudiac/terraform/plugins"       // 预置 providers 目录(己打包到镜像)
	ContainerPluginCachePath = "/cloudiac/terraform/plugins-cache" // terraform plugins 缓存目录
)

const (
	TaskScriptName = "run.sh"
	TaskLogName    = "output.log"

	TaskInfoFileName          = "info.json"
	TaskContainerInfoFileName = "container.json"

	CloudIacTfFile   = "_cloudiac.tf"
	CloudIacPlayVars = "_cloudiac_play_vars.yml"

	TFStateJsonFile  = "tfstate.json"
	TFPlanJsonFile   = "tfplan.json"
	TFProviderSchema = "tfproviderschema.json"

	AnsibleStateAnalysisName = "terraform.py"

	FollowLogDelay = time.Second // follow 文件时读到 EOF 后进行下次读取的等待时长

	PoliciesDir         = "policies"
	TerrascanJsonFile   = "_tfscan.json"
	TerrascanResultFile = "_tfresult.json"
	TerrascanLogFile    = "_tsscan.log"
	RegoResultFile      = "_rego.json"

	PopulateSourceLineCount = 3
)
