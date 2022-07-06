// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package consts

import (
	"cloudiac/common"
	"time"
)

const (
	LowerCaseLetter = "abcdefghijklmnopqrstuvwxyz"
	UpperCaseLetter = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Letter          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	DigitChars      = "0123456789"
	SpecialChars    = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

	DefaultPageSize = 15   // 默认分页大小
	MaxPageSize     = 5000 // 最大单页数据条数

	MaxLogContentSize = 1024 * 1024 // 最大日志文件大小，超限会被截断

	RunnerConnectTimeout = time.Second * 5
	DbTaskPollInterval   = time.Second * 3 // 轮询 db 任务状态的间隔

	DefaultAdminEmail = "admin@example.com"

	CtxKey = "__request_ctx__"

	DemoOrgId = "org-demo0000000000000000"

	SysUserId       = "u-system00000000000000"
	DefaultSysEmail = "sys@example.com"
	DefaultSysName  = "System"

	DefaultTerraformVersion = "0.14.11"

	// token subject
	JwtSubjectUserAuth  = "userAuth" // 用于用户认证
	JwtSubjectSsoCode   = "ssoCode"  // 用于 sso 单点登录
	JwtSubjectActivate  = "activate" // 用于账号激活
	UserEmailINActivate = "inactive" // 用于账号激活
	UserEmailActivate   = "active"   // 用于账号激活

	DirRoot                          = "/"
	PolicyGroupDownloadTimeoutSecond = 20 * time.Second
	PolicySeverityHigh               = "HIGH"
	PolicySeverityMedium             = "MEDIUM"
	PolicySeverityLow                = "LOW"

	RegistryMirrorUri = "/v1/mirrors/providers/"

	AuthRegisterActivationPath = "/activation/"
	AuthPasswordResetPath      = "/find-password/"

	AuthRegisterActivationSubject = "欢迎注册 CloudIaC"
	AuthPasswordResetSubject      = "CloudIaC 密码重置"
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

	GitTypeGitLab   = "gitlab"
	GitTypeGitEA    = "gitea"
	GitTypeGithub   = "github"
	GitTypeGitee    = "gitee"
	GitTypeLocal    = "local"
	GitTypeRegistry = "registry"

	MetaYmlMatch   = "meta.y*ml"
	VariablePrefix = "variables.tf"

	TfVarFileMatch    = "*.tfvars"
	TfFileMatch       = "*.tf"
	TplTfCheckSuccess = "Success"
	TplTfCheckFailed  = "Failed"
	PlaybookDir       = "ansible"
	PlaybookMatch     = "*.y*ml"

	IacTaskLogPrefix = "*** IaC: " // IaC 写入 message 到任务日志时使用的统一前缀

	LocalGitReposPath                            = "repos"    // 内置 http git server 服务目录
	LocalGitReposLocalGitReposPathSubdirectories = "cloudiac" // 内置 http git server 服务子目录
	ReposUrlPrefix                               = "/repos"   // 内置 http git server url prefix

	DefaultVcsName  = "默认仓库"
	RegistryVcsName = "Registry"

	PolicyRego = "*.rego"

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

	EnvAbortManager = ""

	EnvMaxTagLength = 20
	EnvMaxTagNum    = 5

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
	TaskSourceApi          = "api"

	TaskAutoDestroyName = "Auto Destroy"

	BillCollectAli = "alicloud"

	//terraform action type
	TerraformActionCreate = "create"
	TerraformActionUpdate = "update"
	TerraformActionDelete = "delete"

	DemoEnvTTL = "12h"

	TemplateSourceVcs      = "vcs"
	TemplateSourceRegistry = "registry"
)

const (
	AlicloudAK = "ALICLOUD_ACCESS_KEY"
	AlicloudSK = "ALICLOUD_SECRET_KEY"
)

var (
	EnvScopeEnv     = []string{ScopeEnv, ScopeTemplate, ScopeProject, ScopeOrg}
	EnvScopeTpl     = []string{ScopeTemplate, ScopeOrg}
	EnvScopeProject = []string{ScopeProject, ScopeOrg}
	EnvScopeOrg     = []string{ScopeOrg}

	// 按优先级从低到高排序的变量 scopes
	SortedVarScopes = []string{ScopeOrg, ScopeTemplate, ScopeProject, ScopeEnv}

	VariableGroupEnv     = []string{ScopeOrg, ScopeProject, ScopeTemplate, ScopeEnv}
	VariableGroupTpl     = []string{ScopeOrg, ScopeTemplate}
	VariableGroupProject = []string{ScopeOrg, ScopeProject}
	VariableGroupOrg     = []string{ScopeOrg}

	TaskActiveStatus = []string{common.TaskPending, common.TaskRunning, common.TaskApproving}

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

	BillProviderResAccount = map[string][]string{
		BillCollectAli: []string{AlicloudAK, AlicloudSK},
	}
)
