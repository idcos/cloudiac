package forms

import (
	"cloudiac/portal/models"
	"time"
)

type CreatePolicyForm struct {
	BaseForm

	Name          string    `json:"name" binding:"required" example:"ECS分配公网IP"`                                                          // 策略名称
	FixSuggestion string    `json:"fixSuggestion" binding:"" example:"1. 设置 internet_max_bandwidth_out = 0\n 2. 取消设置 allocate_public_ip"` // 修复建议
	Severity      string    `json:"severity" binding:"" enums:"'high','medium','low'" example:"medium"`                                   // 严重性
	Tags          string    `json:"tags" form:"tags" example:"aliyun,jscloud"`
	GroupId       models.Id `json:"groupId" form:"groupId"`

	Rego string `json:"rego"` // rego脚本
}

type SearchPolicyForm struct {
	PageForm

	Q        string      `form:"q" json:"q" binding:""` // 策略组名称，支持模糊搜索
	Severity string      `json:"severity" form:"severity" enums:"'high','medium','low','none'" example:"medium"`
	GroupId  []models.Id `json:"groupId" form:"groupId" `
}

type UpdatePolicyForm struct {
	BaseForm

	Id            models.Id `uri:"id"`
	Name          string    `json:"name" example:"ECS分配公网IP"`                                                                             // 策略名称
	FixSuggestion string    `json:"fixSuggestion" binding:"" example:"1. 设置 internet_max_bandwidth_out = 0\n 2. 取消设置 allocate_public_ip"` // 修复建议
	Severity      string    `json:"severity" binding:"" enums:"'high','medium','low','none'" example:"medium"`                            // 严重性
	Tags          string    `json:"tags" form:"tags" example:"aliyun,jscloud"`
	GroupId       models.Id `json:"groupId" form:"groupId"`

	Rego    string `json:"rego"` // rego脚本
	Enabled bool   `json:"enabled" form:"enabled"`
}

type DeletePolicyForm struct {
	BaseForm

	Id models.Id `uri:"id"`
}

type DetailPolicyForm struct {
	BaseForm

	Id models.Id `uri:"id"`
}

/*OrgId     Id
CreatorId Id `json:"creatorId" gorm:"size:32;not null;comment:创建人ID" example:"u-c3lcrjxczjdywmk0go90"`

Name        string `json:"name" gorm:"not null;size:128;comment:策略组名称" example:"安全合规策略组"`
Description string `json:"description" gorm:"type:text;comment:描述" example:"本组包含对于安全合规的检查策略"`
Enabled     bool   `json:"enabled" gorm:"default:true;comment:是否启用" example:"true"`
Source      string `json:"source" gorm:"type:enum('vcs','registry');comment:来源：VCS/Registry"`
VcsId       Id     `json:"vcsId" gorm:"size:32;not null;comment:VCS ID"`
RepoId      string `json:"repoId" gorm:"size:128;not null;comment:VCS 仓库ID"`
GitTags     string `json:"gitTags" gorm:"size:128;comment:Git 版本标签：\"v1.0.0\""`
Branch      string `json:"branch" gorm:"size:128;comment:分支"`
CommitId    string `json:"commitId" gorm:"size:128;not null;当前 git commit id"`
UseLatest   bool   `json:"useLatest" gorm:"default:false;comment:是否跟踪最新版本，如果从分支导入，默认为true" example:"true"`
Version     string `json:"version" gorm:"size:32;not null;策略组版本：\"1.0.0\""`
Dir         string `json:"dir" gorm:"default:\"/\";comment:策略组目录，默认为根目录：/"`
Label       string `json:"label" gorm:"size:128;comment:策略组标签，多个值以 , 分隔"`*/

type CreatePolicyGroupForm struct {
	BaseForm

	Name        string `json:"name" binding:"required" example:"安全合规策略组"`
	Description string `json:"description" binding:"" example:"本组包含对于安全合规的检查策略"`
	Label       string `json:"label" binding:"" example:"security,alicloud"`

	Source  string    `json:"source" binding:"required" enums:"vcs,registry" example:"来源"`
	VcsId   models.Id `json:"vcsId" binding:"required" example:"vcs-c3lcrjxczjdywmk0go90"`
	RepoId  string    `json:"repoId" binding:"required" example:"1234567890"`
	GitTags string    `json:"gitTags" example:"Git Tags"`
	Branch  string    `json:"branch" example:"master"`
	Dir     string    `json:"dir" example:"/"`
}

type SearchPolicyGroupForm struct {
	NoPageSizeForm

	Q string `form:"q" json:"q" binding:""` // 策略组名称，支持模糊搜索
}

type UpdatePolicyGroupForm struct {
	BaseForm

	Id          models.Id `uri:"id"`
	Name        string    `json:"name" form:"name" `
	Description string    `json:"description" binding:"" example:"本组包含对于安全合规的检查策略"`
	Enabled     bool      `json:"enabled" form:"enabled"`

	Label       string `json:"label" binding:"" example:"security,alicloud"`
	Source  string    `json:"source" binding:"required" enums:"vcs,registry" example:"来源"`
	VcsId   models.Id `json:"vcsId" binding:"required" example:"vcs-c3lcrjxczjdywmk0go90"`
	RepoId  string    `json:"repoId" binding:"required" example:"1234567890"`
	GitTags string    `json:"gitTags" example:"Git Tags"`
	Branch  string    `json:"branch" example:"master"`
	Dir     string    `json:"dir" example:"/"`
}

type DeletePolicyGroupForm struct {
	BaseForm

	Id models.Id `uri:"id"`
}

type DetailPolicyGroupForm struct {
	BaseForm

	Id models.Id `uri:"id"`
}

type UpdatePolicyRelForm struct {
	BaseForm

	Id             models.Id   `uri:"id" binding:"" example:"tpl-c3ek0co6n88ldvq1n6ag"`
	PolicyGroupIds []models.Id `json:"policyGroupIds" binding:"required" example:"pog-c3ek0co6n88ldvq1n6ag,pog-c3ek0co6n88ldvq1n6bg"`
	Scope          string      `json:"-" swaggerignore:"true" binding:""`
}

type EnableScanForm struct {
	BaseForm

	Id      models.Id `uri:"id" swaggerignore:"true" binding:"" example:"tpl-c3ek0co6n88ldvq1n6ag"` // ID
	Scope   string    `json:"-" swaggerignore:"true" binding:""`
	Enabled bool      `json:"enabled" binding:"" example:"true"` // 是否启用扫描
}

type ScanTemplateForm struct {
	BaseForm

	Id    models.Id `uri:"id" binding:"" example:"tpl-c3ek0co6n88ldvq1n6ag"`      // 云模板Id
	Parse bool      `json:"parse" binding:""  enums:"true,false" example:"false"` // 是否只执行解析
}

type ScanEnvironmentForm struct {
	BaseForm

	Id    models.Id `uri:"id" binding:"" example:"env-c3ek0co6n88ldvq1n6ag"`      // 环境Id
	Parse bool      `json:"parse" binding:""  enums:"true,false" example:"false"` // 是否只执行解析
}

type PolicyParseForm struct {
	BaseForm

	TemplateId models.Id `form:"tplId" json:"tplId" binding:"" example:"tpl-c3ek0co6n88ldvq1n6ag"` // 云模板Id
	EnvId      models.Id `form:"envId" json:"envId" binding:"" example:"env-c3ek0co6n88ldvq1n6ag"` // 云模板Id
}

type OpnPolicyAndPolicyGroupRelForm struct {
	BaseForm

	PolicyGroupId models.Id `uri:"id" json:"policyGroupId" form:"policyGroupId" `
	RmPolicyIds   []string  `json:"rmPolicyIds" binding:"" example:"po-c3ek0co6n88ldvq1n6ag"`
	AddPolicyIds  []string  `json:"addPolicyIds" binding:"" example:"po-c3ek0co6n88ldvq1n6ag"`
}

type CreatePolicySuppressForm struct {
	BaseForm

	CreatorId models.Id   `json:"creatorId" `
	TplId     models.Id   `json:"tplId" `
	EnvId     []models.Id `json:"envId" `
	PolicyId  models.Id   `json:"policyId" `

	Reason string `json:"reason" form:"reason" `
	Type   string `json:"type" form:"type" enums:"'strategy','source'"`
}

type SearchPolicySuppressForm struct {
	PageForm

	Id models.Id `uri:"id"`
}

type SearchPolicySuppressSourceForm struct {
	NoPageSizeForm

	Id models.Id `uri:"id"`
}

type DeletePolicySuppressForm struct {
	BaseForm

	Id         models.Id `uri:"id"`
	SuppressId models.Id `uri:"suppressId"`
}

type SearchPolicyTplForm struct {
	NoPageSizeForm

	OrgId models.Id `form:"orgId" binding:""` // 组织ID
	TplId models.Id `form:"tplId" binding:""`
	Q     string    `form:"q" json:"q" binding:""` // 模糊搜索
}

type DetailPolicyTplForm struct {
	BaseForm
	Id models.Id `json:"id" form:"id" `
}

type TplOfPolicyForm struct {
	PageForm

	Id       models.Id `json:"id" form:"id" `
	Q        string    `form:"q" json:"q" binding:""` // 策略组名称，支持模糊搜索
	Severity string    `json:"severity" form:"severity" enums:"'high','medium','low','none'" example:"medium"`
	GroupId  models.Id `json:"groupId" form:"groupId" `
}

type TplOfPolicyGroupForm struct {
	NoPageSizeForm

	Id models.Id `uri:"id"`
}

type SearchPolicyEnvForm struct {
	NoPageSizeForm

	OrgId     models.Id `form:"orgId" binding:""`
	ProjectId models.Id `form:"projectId" binding:""`
	EnvId     models.Id `form:"envId" binding:""`
	Q         string    `form:"q" json:"q" binding:""` // 模糊搜索
}

type EnvOfPolicyForm struct {
	PageForm

	Id       models.Id `json:"id" form:"id" `
	Q        string    `form:"q" json:"q" binding:""` // 策略组名称，支持模糊搜索
	Severity string    `json:"severity" form:"severity" enums:"'high','medium','low','none'" example:"medium"`
	GroupId  models.Id `json:"groupId" form:"groupId" `
}

type PolicyErrorForm struct {
	PageForm
	Id models.Id `uri:"id"`
	Q  string    `json:"q" form:"q"`
}

type UpdatePolicySuppressForm struct {
	BaseForm
	Id           models.Id   `uri:"id" swaggerignore:"true"`                          // 策略ID
	Reason       string      `json:"reason" example:"测试环境无需检测"`                       // 屏蔽原因
	AddSourceIds []models.Id `json:"addTargetIds" example:"env-c3ek0co6n88ldvq1n6ag"` // 添加屏蔽源ID列表
	//RmSourceIds  []models.Id `json:"rmTargetIds" example:"env-c3ek0co6n88ldvq1n6ag"`  // 删除屏蔽源ID列表
}

type PolicyScanResultForm struct {
	NoPageSizeForm

	Id models.Id `uri:"id" `
}

type PolicyScanReportForm struct {
	BaseForm

	Id        models.Id `uri:"id" swaggerignore:"true"`                               // 策略/策略组ID
	From      time.Time `json:"from" form:"from" example:"2006-01-02T15:04:05Z07:00"` //  开始日期
	To        time.Time `json:"to" form:"to" example:"2006-01-02T15:04:05Z07:00"`     // 结束日期
	ShowCount int       `json:"showCount" form:"showCount" example:"5"`
}

type PolicyTestForm struct {
	BaseForm

	Input string `form:"input" json:"input" binding:"" example:"{\n\"alicloud_instance\": [\n\n{\t\n\"id\": \"alicloud_instance.instance\"..."` // 脚本验证源数据
	Rego  string `form:"rego" json:"rego" binding:"" example:"package accurics\ninstanceWithNoVpc[retVal] {..."`                                // rego脚本内容
}

type PolicyLastTasksForm struct {
	PageForm

	Id    models.Id `uri:"id" `
	Scope string    `json:"-"`
}

type SearchGroupOfPolicyForm struct {
	PageForm

	Id     models.Id `uri:"id" `
	IsBind bool      `json:"bind" form:"bind" ` //  ture: 查询绑定策略组的策略，false: 查询未绑定的策略组的策略
}
