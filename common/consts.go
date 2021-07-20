package common

import "cloudiac/portal/consts"

// 以下值会在编译时动态注入(见 Makefile)。
var (
	VERSION = "v0.0.0"
	BUILD   = "000000"
)

const (
	TaskTypePlan    = "plan"    // 计划执行，不会修改资源或做服务配置
	TaskTypeApply   = "apply"   // 执行 terraform apply 和 playbook
	TaskTypeDestroy = "destroy" // 销毁，删除所有资源

	TaskPending   = "pending"
	TaskRunning   = "running"
	TaskApproving = "approving"
	TaskFailed    = "failed"
	TaskComplete  = "complete"
	//TaskTimeout   = "timeout"

	TaskStepInit    = "init"
	TaskStepPlan    = "plan"
	TaskStepApply   = "apply"
	TaskStepDestroy = "destroy"
	TaskStepPlay    = "play"    // play playbook
	TaskStepCommand = "command" // run command

	TaskStepPending   = "pending"
	TaskStepApproving = "approving"
	TaskStepRejected  = "rejected"
	TaskStepRunning   = "running"
	TaskStepFailed    = "failed"
	TaskStepComplete  = "complete"
	TaskStepTimeout   = "timeout"

	TaskTypePlanName    = "plan"
	TaskTypeApplyName   = "apply"
	TaskTypeDestroyName = "destroy"

	TaskStepTimeoutDuration = 600
)

var (
	EnvScopeEnv     = []string{consts.ScopeEnv, consts.ScopeTemplate, consts.ScopeProject, consts.ScopeOrg}
	EnvScopeTpl     = []string{consts.ScopeTemplate, consts.ScopeOrg}
	EnvScopeProject = []string{consts.ScopeProject, consts.ScopeOrg}
	EnvScopeOrg     = []string{consts.ScopeOrg}
)
