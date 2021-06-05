package consts

import (
	"time"
)

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

	DefaultKeepalive               = 1 * time.Second
	DefaultCollectTaskSyncInterval = time.Minute
	MaxDsReportMetricsInterval     = 10 * time.Minute

	// 当前代码对应的数据库表结构版本号。
	// 大部分情况下修改 model 会自动执行 migrate，但某些情况下我们的 migrate 只需要针对特定版本执行，
	// 所以我们通过在 db 中记录 scheme 版本号来实现。
	// 在每次执行完 migrate 后 db 中记录的 scheme 版本都会更新为下面的值，
	// 这样特殊的 migrate 操作只需要加上版本判断就能满足需求
	DBSchemeVersion = "20201108"

	MS        = "MONITOR_SOURCE"
	MM        = "MONITOR_METRIC"
	RL        = "RELATION_LIST"
	CreateMS  = "CREATE_MONITOR_SOURCE"
	ModifyMS  = "MODEFY_MONITOR_SOURCE"
	DeleteMS  = "DELETE_MONITOR_SOURCE"
	DisableMS = "DISABLE_MONITOR_SOURCE"
	EnableMS  = "ENABLE_MONITOR_SOURCE"
	CreateMM  = "CREATE_MONITOR_METRIC"
	ModifyMM  = "MODEFY_MONITOR_METRIC"
	DeleteMM  = "DELETE_MONITOR_METRIC"
	ExportRL  = "EXPORT_RELATION_LIST"
	ImportRL  = "IMPORT_RELATION_LIST"
)

const (
	MetricCatCpu   = "cpu"
	MetricCatMem   = "mem"
	MetricCatDisk  = "disk"
	MetricCatNet   = "net"
	MetricCatDb    = "db"
	MetricCatApp   = "app"
	MetricCatMw    = "mw" // middleware
	MetricCatOther = "other"

	MetircLabelResId  = "opsnow_res"
	MetircLabelDsName = "opsnow_ds"

	DsTypeZabbix = "zabbix"
	DsTypeProm   = "prom"

	//作业状态
	TaskPending   = "pending"
	TaskAssigning = "assigning"

	TaskRunning  = "running"
	TaskTimeout  = "timeout"
	TaskFailed   = "failed"
	TaskComplete = "complete"
	TaskLogName  = "runner.log"
	TaskApply    = "apply"
	TaskPlan     = "plan"

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
