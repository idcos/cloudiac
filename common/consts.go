// Copyright 2021 CloudJ Company Limited. All rights reserved.

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
	TaskTypeScan    = "scan"    // 策略扫描，只执行策略扫描，不锈钢资源或配置

	TaskPending   = "pending"
	TaskRunning   = "running"
	TaskApproving = "approving"
	TaskFailed    = "failed"
	TaskComplete  = "complete"

	TaskStepInit    = "init"
	TaskStepPlan    = "plan"
	TaskStepApply   = "apply"
	TaskStepDestroy = "destroy"
	TaskStepPlay    = "play"    // play playbook
	TaskStepCommand = "command" // run command
	TaskStepCollect = "collect" // 任务结束后的信息采集
	TaskStepTfParse = "tfparse" // 云模板解析
	TaskStepTfScan  = "tfscan"  // 云模板策略扫描

	CollectTaskStepIndex = -1

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
