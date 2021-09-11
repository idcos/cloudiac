// Copyright 2021 CloudJ Company Limited. All rights reserved.

package common

// 以下值会在编译时动态注入(见 Makefile)。
var (
	VERSION = "v0.0.0"
	BUILD   = "000000"
)

const (
	TaskTypePlan    = "plan"    // 计划执行，不会修改资源或做服务配置
	TaskTypeApply   = "apply"   // 执行 terraform apply 和 playbook
	TaskTypeDestroy = "destroy" // 销毁，删除所有资源
	TaskTypeScan    = "scan"    // 策略扫描，只执行策略扫描，不锈钢资源或配置
	TaskTypeParse   = "parse"   // 策略扫描，只执行策略扫描，不锈钢资源或配置

	TaskPending   = "pending"
	TaskRunning   = "running"
	TaskApproving = "approving"
	TaskRejected  = "rejected"
	TaskFailed    = "failed"
	TaskComplete  = "complete"

	TaskStepInit     = "init"
	TaskStepPlan     = "plan"
	TaskStepApply    = "apply"
	TaskStepDestroy  = "destroy"
	TaskStepPlay     = "play"    // play playbook
	TaskStepCommand  = "command" // run command
	TaskStepCollect  = "collect" // 任务结束后的信息采集
	TaskStepScanInit = "scaninit"
	TaskStepTfParse  = "tfparse" // 云模板解析
	TaskStepTfScan   = "tfscan"  // 云模板策略扫描

	CollectTaskStepIndex = -1

	TaskStepPending   = "pending"
	TaskStepApproving = "approving"
	TaskStepRejected  = "rejected"
	TaskStepRunning   = "running"
	TaskStepFailed    = "failed"
	TaskStepComplete  = "complete"
	TaskStepTimeout   = "timeout"

	TaskStepPolicyViolationExitCode = 3 // 合规检查不通过时的退出码

	TaskTypePlanName    = "plan"
	TaskTypeApplyName   = "apply"
	TaskTypeDestroyName = "destroy"
	TaskTypeScanName    = "scan"

	TaskStepTimeoutDuration = 600

	VcsGitlab = "gitlab"
	VcsGitea  = "gitea"
	VcsGitee  = "gitee"
	VcsGithub = "github"

	PolicyStatusPending    = "pending"
	PolicyStatusPassed     = "passed"
	PolicyStatusFailed     = "failed"
	PolicyStatusViolated   = "violated"
	PolicyStatusSuppressed = "suppressed"

	PolicySeverityHigh   = "high"
	PolicySeverityMedium = "medium"
	PolicySeverityLow    = "low"

	PolicySuppressTypeSource = "source"
	PolicySuppressTypePolicy = "policy"
)

var (
	TerraformVersions = []string{
		"0.11.15",
		"0.12.31",
		"0.13.7",
		"0.14.11",
		"0.15.5",
		"1.0.6",
	}
)
