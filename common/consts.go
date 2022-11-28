// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package common

// 以下值会在编译时动态注入(见 Makefile)。
var (
	VERSION = "v0.0.0"
	BUILD   = "000000"
)

const (
	TaskTypePlan     = "plan"     // 计划执行，不会修改资源或做服务配置
	TaskTypeApply    = "apply"    // 执行 terraform apply 和 playbook
	TaskTypeDestroy  = "destroy"  // 销毁，删除所有资源
	TaskTypeScan     = "scan"     // 策略扫描，只执行策略扫描，不修改资源或配置
	TaskTypeParse    = "parse"    // 策略扫描，只执行策略扫描，不修改资源或配置
	TaskTypeEnvScan  = "envScan"  // 环境策略扫描，只执行策略扫描，不修改资源或配置
	TaskTypeEnvParse = "envParse" // 环境策略扫描，只执行策略扫描，不修改资源或配置
	TaskTypeTplScan  = "tplScan"  // 云模板策略扫描，只执行策略扫描，不修改资源或配置
	TaskTypeTplParse = "tplParse" // 云模板策略扫描，只执行策略扫描，不修改资源或配置

	// TODO 与 taskTypexxx 重复，需要替换
	TaskJobPlan     = "plan"
	TaskJobApply    = "apply"
	TaskJobDestroy  = "destroy"
	TaskJobScan     = "scan"
	TaskJobParse    = "parse"
	TaskJobEnvScan  = "envScan"
	TaskJobEnvParse = "envParse"
	TaskJobTplScan  = "tplScan"
	TaskJobTplParse = "tplParse"

	TaskPending   = "pending"
	TaskRunning   = "running"
	TaskApproving = "approving"
	TaskRejected  = "rejected"
	TaskFailed    = "failed"
	TaskAborted   = "aborted"
	TaskComplete  = "complete"

	TaskStepCheckout  = "checkout"
	TaskStepTfInit    = "terraformInit"
	TaskStepTfPlan    = "terraformPlan"
	TaskStepTfApply   = "terraformApply"
	TaskStepTfDestroy = "terraformDestroy"

	// 0.3 扫描步骤名称
	TaskStepOpaScan = "opaScan" // 云模板策略扫描
	// 0.4 扫描步骤名称
	TaskStepTplParse = "tplParse"
	TaskStepTplScan  = "tplScan"
	TaskStepEnvParse = "envParse"
	TaskStepEnvScan  = "envScan"

	TaskStepAnsiblePlay = "ansiblePlay" // play playbook
	TaskStepCommand     = "command"     // run command
	TaskStepCollect     = "collect"     // 任务结束后的信息采集
	TaskStepScanInit    = "scaninit"
	CronDriftTaskName   = "Drift Detection" // 漂移检测任务名称

	PipelineFileName = ".cloudiac-pipeline.yml"

	// 结束采集步骤的索引
	CollectTaskStepIndex = -1

	TaskStepPending   = "pending"
	TaskStepApproving = "approving"
	TaskStepRejected  = "rejected"
	TaskStepRunning   = "running"
	TaskStepFailed    = "failed"
	TaskStepComplete  = "complete"
	TaskStepTimeout   = "timeout"
	TaskStepAborted   = "aborted"

	TaskStepPolicyViolationExitCode = 3 // 合规检查不通过时的退出码

	TaskTypePlanName     = "plan"
	TaskTypeApplyName    = "apply"
	TaskTypeDestroyName  = "destroy"
	TaskTypeScanName     = "scan"
	TaskTypeEnvScanName  = "envScan"
	TaskTypeEnvParseName = "envParse"
	TaskTypeTplScanName  = "tplScan"
	TaskTypeTplParseName = "tplParse"

	ProjectStatusEnable  = "enable"
	ProjectStatusDisable = "disable"

	// 默认步骤超时时间(秒)
	DefaultTaskStepTimeoutSecond = 3600

	VcsGitlab = "gitlab"
	VcsGitea  = "gitea"
	VcsGitee  = "gitee"
	VcsGithub = "github"

	// PolicyStatusPending 检测中
	PolicyStatusPending = "pending"
	// PolicyStatusPassed 通过
	PolicyStatusPassed = "passed"
	PolicyStatusFailed = "failed"
	// PolicyStatusViolated 不通过
	PolicyStatusViolated = "violated"
	// PolicyStatusSuppressed 屏蔽
	PolicyStatusSuppressed = "suppressed"
	//PolicyStatusEnable 未检测
	PolicyStatusEnable = "enable"
	//PolicyStatusDisable 未开启
	PolicyStatusDisable = "disable"

	PolicySeverityHigh   = "high"
	PolicySeverityMedium = "medium"
	PolicySeverityLow    = "low"

	PolicySuppressTypeSource = "source"
	PolicySuppressTypePolicy = "policy"

	RunnerServiceName    = "CT-Runner"
	IacPortalServiceName = "IaC-Portal"

	ConsulCa            = "ca.pem"
	ConsulCakey         = "client.key"
	ConsulCapem         = "client.pem"
	ConsulContainerPath = "/cloudiac/cert/"
	ConsulSessionTTL    = 10
)

var (
	TerraformVersions = []string{
		// 以下两个版本不支持 network mirror 协议。
		// 这里注释掉，但 worker 镜像里还是会安装，以保证兼容性
		// "0.11.15",
		// "0.12.31",

		"0.13.7",
		"0.14.11",
		"0.15.5",
		"1.0.6",
		"1.1.9",
		"1.2.4",
	}
)
