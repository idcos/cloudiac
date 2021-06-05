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
	TaskRunning   = "running"
	TaskTimeout   = "timeout"
	TaskFailed    = "failed"
	TaskComplete  = "complete"

	TaskLogName = "runner.log"
	TaskApply   = "apply"
	TaskPlan    = "plan"

	ResourceAccountDisable = "disable"
	ResourceAccountEnable  = "enable"
	DockerStatusExited     = "exited"
	Terraform              = "terraform"
	TerraformVar           = "TF_VAR_"
	WorkFlow               = "workflow"
	GitLab                 = "gitlab"
	GitEA                  = "gitea"

	TfVarFileExt    = ".tfvars"
	PlaybookPrefixYml = ".yml"
	PlaybookPrefixYaml = ".yaml"
	IacTaskLogPrefix = "*** IaC: " // IaC 写入 message 到任务日志时使用的统一前缀
)

var (
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
