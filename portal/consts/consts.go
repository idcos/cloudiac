// Copyright 2021 CloudJ Company Limited. All rights reserved.

package consts

import (
	"cloudiac/common"
	"time"
)

const (
	LowerCaseLetter = "abcdefghijklmnopqrstuvwxyz"
	UpperCaseLetter = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	DigitChars      = "0123456789"

	DefaultPageSize = 15   // 默认分页大小
	MaxPageSize     = 5000 // 最大单页数据条数

	MaxLogContentSize = 1024 * 1024 // 最大日志文件大小，超限会被截断

	RunnerConnectTimeout = time.Second * 5
	DbTaskPollInterval   = time.Second // 轮询 db 任务状态的间隔

	DefaultAdminEmail = "admin@example.com"

	CtxKey = "__request_ctx__"

	DemoOrgId = "org-demo0000000000000000"

	SysUserId       = "u-system00000000000000"
	DefaultSysEmail = "sys@example.com"
	DefaultSysName  = "System"

	DefaultTerraformVersion = "0.14.11"

	DirRoot                          = "/"
	PolicyGroupDownloadTimeoutSecond = 20 * time.Second
	PolicySeverityHigh               = "HIGH"
	PolicySeverityMedium             = "MEDIUM"
	PolicySeverityLow                = "LOW"
)

const (
	// 作业状态
	TaskLogName = "runner.log"

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
	VariablePrefix = "variables.tf"

	TfVarFileMatch    = "*.tfvars"
	TplTfCheck        = "*.tf"
	TplTfCheckSuccess = "Success"
	TplTfCheckFailed  = "Failed"
	PlaybookMatch     = "*.y*ml"
	Ansible           = "ansible"

	IacTaskLogPrefix = "*** IaC: " // IaC 写入 message 到任务日志时使用的统一前缀

	LocalGitReposPath = "repos"  // 内置 http git server 服务目录
	ReposUrlPrefix    = "/repos" // 内置 http git server url prefix

	NotificationMessageTitle = "CloudIaC平台系统通知"

	GraphDimensionModule   = "module"
	GraphDimensionProvider = "provider"
	GraphDimensionType     = "type"
)

const (
	SuperAdmin = "root"

	RoleRoot      = "root"
	RoleLogin     = "login"
	RoleAnonymous = "anonymous"
	RoleDemo      = "demo"

	OrgRoleAdmin  = "admin"
	OrgRoleMember = "member"

	ProjectRoleManager  = "manager"  //
	ProjectRoleApprover = "approver" // 要以创建模板、环境，部署审批
	ProjectRoleOperator = "operator" // 可以发起 plan、apply
	ProjectRoleGuest    = "guest"    // 访客，只读权限

	ScopeOrg      = "org"
	ScopeProject  = "project"
	ScopeTemplate = "template"
	ScopeEnv      = "env"

	ScopePolicy      = "policy"
	ScopePolicyGroup = "policyGroup"
	ScopeTask        = "task"

	VarTypeEnv       = "environment"
	VarTypeTerraform = "terraform"
	VarTypeAnsible   = "ansible"

	TokenApi     = "api"     //token类型
	TokenTrigger = "trigger" //token类型

	EnvTriggerPRMR   = "prmr"
	EnvTriggerCommit = "commit"

	EventTaskFailed    = "task.failed"
	EventTaskComplete  = "task.complete"
	EventTaskRunning   = "task.running"
	EventTaskApproving = "task.approving"
	EventTaskRejected  = "task.rejected"
	EvenvtCronDrift    = "task.crondrift"

	DefaultTfMirror   = "https://releases.hashicorp.com/terraform"
	HttpClientTimeout = 20

	TaskCallbackKafka = "kafka"

	TaskSourceManual       = "manual"
	TaskSourceDriftPlan    = "driftPlan"
	TaskSourceDriftApply   = "driftApply"
	TaskSourceWebhookPlan  = "webhookPlan"
	TaskSourceWebhookApply = "webhookApply"
	TaskSourceAutoDestroy  = "autoDestroy"
)

var (
	EnvScopeEnv     = []string{ScopeEnv, ScopeTemplate, ScopeProject, ScopeOrg}
	EnvScopeTpl     = []string{ScopeTemplate, ScopeOrg}
	EnvScopeProject = []string{ScopeProject, ScopeOrg}
	EnvScopeOrg     = []string{ScopeOrg}

	VariableGroupEnv     = []string{ScopeOrg, ScopeProject, ScopeTemplate, ScopeEnv}
	VariableGroupTpl     = []string{ScopeOrg, ScopeTemplate}
	VariableGroupProject = []string{ScopeOrg, ScopeProject}
	VariableGroupOrg     = []string{ScopeOrg}

	StatusTranslation = map[string]string{
		"complete": "成功",
		"failed":   "失败",
		"running":  "运行中",
		"timeout":  "超时",
		"pending":  "排队中",
	}
	TerraformVersions = []string{
		"0.11.15",
		"0.12.31",
		"0.13.7",
		"0.14.11",
		"0.15.5",
		"1.0.6",
	}

	TaskStatusToEventType = map[string]string{
		common.TaskComplete:  EventTaskComplete,
		common.TaskFailed:    EventTaskFailed,
		common.TaskRunning:   EventTaskRunning,
		common.TaskApproving: EventTaskApproving,
		common.TaskRejected:  EventTaskFailed,
		EvenvtCronDrift:      EvenvtCronDrift,
	}
)
