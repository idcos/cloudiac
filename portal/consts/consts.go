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

	DefaultPageSize = 15 // 默认分页大小
	MaxPageSize     = 5000 // 同时是 csv 最大导出条数

	MaxLogContentSize = 1024 * 1024 // 最大日志文件大小，超限会被截断

	DefaultAdminEmail = "admin@example.com"
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

	TfVarFileMatch = "*.tfvars"
	PlaybookMatch  = "*.y*ml"
	Ansible        = "ansible"

	TerraformStateListName = "state_list.log" //terraform state list 文件名称

	IacTaskLogPrefix = "*** IaC: " // IaC 写入 message 到任务日志时使用的统一前缀

	LocalGitReposPath = "repos"  // 内置 http git server 服务目录
	ReposUrlPrefix    = "/repos" // 内置 http git server url prefix
)

const (
	OrgRoleOwner  = "owner"
	OrgRoleMember = "member"

	ProjectRoleOwner    = "owner"    //
	ProjectRoleManager  = "manager"  // 要以创建模板、环境，部署审批
	ProjectRoleOperator = "operator" // 可以发起 plan、apply
	ProjectRoleGuest    = "guest"    // 访客，只读权限

	ScopeOrg      = "org"
	ScopeProject  = "project"
	ScopeTemplate = "template"
	ScopeEnv      = "env"

	VarTypeEnv       = "environment"
	VarTypeTerraform = "terraform"
	VarTypeAnsible   = "ansible"
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
