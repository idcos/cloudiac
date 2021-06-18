package consts

// 以下值会在编译时动态注入(见 Makefile)。
var (
	VERSION = "v0.0.0"
	BUILD   = "000000"
)

const (
	LowerCaseLetter = "abcdefghijklmnopqrstuvwxyz"
	UpperCaseLetter = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	DigitChars      = "0123456789"

	DefaultPageSize = 15
	MaxPageSize     = 5000 // 同时是 csv 最大导出条数

	DefaultAdminEmail = "admin@example.com"
)

const (
	// 作业状态

	TaskPending   = "pending"
	TaskAssigning = "assigning"

	TaskRunning  = "running"
	TaskTimeout  = "timeout"
	TaskFailed   = "failed"
	TaskComplete = "complete"

	TaskLogName = "runner.log"
	TaskApply   = "apply"
	TaskPlan    = "plan"
	TaskDestroy = "destroy"

	ResourceAccountDisable = "disable"
	ResourceAccountEnable  = "enable"
	DockerStatusExited     = "exited"
	Terraform              = "terraform"
	Env                    = "env"
	TerraformVar           = "TF_VAR_"
	WorkFlow               = "workflow"

	GitTypeGitLab = "gitlab"
	GitTypeGitEA  = "gitea"
	GitTypeGithub = "github"
	GitTypeGitee  = "gitee"
	GitTypeLocal  = "local"

	MetaYmlMatch   = "meta.y*ml"
	VariablePrefix = "variable.tf"

	TfVarFileMatch = "*.tfvars"
	PlaybookMatch  = "playbook.y*ml"

	TerraformStateListName = "state_list.log" //terraform state list 文件名称

	IacTaskLogPrefix = "*** IaC: " // IaC 写入 message 到任务日志时使用的统一前缀

	LocalGitReposPath = "repos"  // 内置 http git server 服务目录
	ReposUrlPrefix    = "/repos" // 内置 http git server url prefix
)

var (
	BomUtf8    = []byte{0xEF, 0xBB, 0xBF}
	AccountMap = map[string]map[string]string{
		"aliyun": {
			"accessKeyId":     "ALICLOUD_ACCESS_KEY",
			"secretAccessKey": "ALICLOUD_SECRET_KEY",
		},
		"vmware": {
			"userName": "username",
			"password": "password",
		},
		"huawei": {},
	}
	StatusTranslation = map[string]string{
		"complete": "成功",
		"failed":   "失败",
		"running":  "运行中",
		"timeout":  "超时",
		"pending":  "排队中",
	}
)
