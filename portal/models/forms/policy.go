package forms

import (
	"cloudiac/portal/models"
	"time"
)

type CreatePolicyForm struct {
	BaseForm

	Name          string `json:"name" binding:"required" example:"ECS分配公网IP"`                                                          // 策略名称
	FixSuggestion string `json:"fixSuggestion" binding:"" example:"1. 设置 internet_max_bandwidth_out = 0\n 2. 取消设置 allocate_public_ip"` // 修复建议
	Severity      string `json:"severity" binding:"" enums:"'high','medium','low'" example:"medium"`                                   // 严重性
	Tags          string `json:"tags" form:"tags" example:"aliyun,jscloud"`

	Rego string `json:"rego" binding:"required"` // rego脚本
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

type CreatePolicyGroupForm struct {
	BaseForm

	Name        string `json:"name" binding:"required" example:"安全合规策略组"`
	Description string `json:"description" binding:"" example:"本组包含对于安全合规的检查策略"`

	//PolicyIds []string `json:"policyIds" binding:"" example:"[\"po-c3ek0co6n88ldvq1n6ag\"]"`
}

type SearchPolicyGroupForm struct {
	PageForm

	Q string `form:"q" json:"q" binding:""` // 策略组名称，支持模糊搜索
}

type UpdatePolicyGroupForm struct {
	BaseForm

	Id          models.Id `uri:"id"`
	Name        string    `json:"name" form:"name" `
	Description string    `json:"description" binding:"" example:"本组包含对于安全合规的检查策略"`
	Enabled     bool      `json:"enabled" form:"enabled"`
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
	PageForm

	Id models.Id `uri:"id"`
}

type DeletePolicySuppressForm struct {
	BaseForm

	Id         models.Id `uri:"id"`
	SuppressId models.Id `uri:"suppressId"`
}

type SearchPolicyTplForm struct {
	PageForm

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

type SearchPolicyEnvForm struct {
	PageForm

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
	Id           models.Id   `uri:"id"`
	Scope        string      `json:"scope" example:"source" enums:"source,policy"`    // 屏蔽类型，source按来源屏蔽，policy屏蔽此策略
	Reason       string      `json:"reason" example:"测试环境无需检测"`                       // 屏蔽原因
	AddSourceIds []models.Id `json:"addTargetIds" example:"env-c3ek0co6n88ldvq1n6ag"` // 添加屏蔽源ID列表
	//RmSourceIds  []models.Id `json:"rmTargetIds" example:"env-c3ek0co6n88ldvq1n6ag"`  // 删除屏蔽源ID列表
}

type PolicyScanResultForm struct {
	PageForm

	Id    models.Id `uri:"id" `
	Scope string    `json:"-"`
}

type PolicyScanReportForm struct {
	BaseForm

	Id    models.Id `uri:"id" `
	Scope string    `json:"-"`
	From  time.Time `json:"from" form:"from" example:"2006-01-02T15:04:05Z07:00"` //  开始日期
	To    time.Time `json:"to" form:"to" example:"2006-01-02T15:04:05Z07:00"`     // 结束日期
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
