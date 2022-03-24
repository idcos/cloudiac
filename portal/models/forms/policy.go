// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import (
	"cloudiac/portal/models"
	"time"
)

type CreatePolicyForm struct {
	BaseForm

	Name          string    `json:"name" binding:"required,gte=2,lte=255" example:"ECS分配公网IP"`                                                   // 策略名称
	FixSuggestion string    `json:"fixSuggestion" binding:"max=255" example:"1. 设置 internet_max_bandwidth_out = 0\n 2. 取消设置 allocate_public_ip"` // 修复建议
	Severity      string    `json:"severity" binding:"required,oneof=high medium low" enums:"'high','medium','low'" example:"medium"`            // 严重性
	Tags          string    `json:"tags" form:"tags" binding:"max=255" example:"aliyun,jscloud"`
	GroupId       models.Id `json:"groupId" form:"groupId" binding:"required,startswith=pog-,max=32"`

	Rego string `json:"rego" binding:"max=255"` // rego脚本
}

type SearchPolicyForm struct {
	PageForm
	Q        string      `form:"q" json:"q" binding:""`                                                                                                         // 策略组名称，支持模糊搜索
	Severity string      `json:"severity" form:"severity" binding:"omitempty,oneof=high medium low none" enums:"'high','medium','low','none'" example:"medium"` //严重性
	GroupId  []models.Id `json:"groupId" form:"groupId" binding:"omitempty,dive,required,startswith=pog-,max=32"`                                               //组织ID
}

type UpdatePolicyForm struct {
	BaseForm

	Id            models.Id `uri:"id" binding:"required,startswith=po-,max=32"`
	Name          string    `json:"name" binding:"omitempty,gte=2,lte=255" example:"ECS分配公网IP"`                                                  // 策略名称
	FixSuggestion string    `json:"fixSuggestion" binding:"max=255" example:"1. 设置 internet_max_bandwidth_out = 0\n 2. 取消设置 allocate_public_ip"` // 修复建议
	Severity      string    `json:"severity" binding:"omitempty,oneof=high medium low" enums:"'high','medium','low','none'" example:"medium"`    // 严重性
	Tags          string    `json:"tags" form:"tags" binding:"max=255" example:"aliyun,jscloud"`
	GroupId       models.Id `json:"groupId" form:"groupId" binding:"omitempty,dive,required,startswith=org-"`

	Rego    string `json:"rego" binding:"max=255"` // rego脚本
	Enabled bool   `json:"enabled" form:"enabled" binding:""`
}

type DeletePolicyForm struct {
	BaseForm

	Id models.Id `uri:"id" binding:"required,startswith=po-,max=32"`
}

type DetailPolicyForm struct {
	BaseForm
	Id models.Id `uri:"id" binding:"required,startswith=po-,max=32" swaggerignore:"true"`
}

type CreatePolicyGroupForm struct {
	BaseForm

	Name        string   `json:"name" binding:"required,gte=2,lte=64" example:"安全合规策略组"`
	Description string   `json:"description" binding:"lte=255" example:"本组包含对于安全合规的检查策略"`
	Labels      []string `json:"labels" binding:"omitempty,dive,required,max=128" example:"[security,alicloud]"`

	Source  string    `json:"source" binding:"required,oneof=vcs registry" enums:"vcs,registry" example:"来源"`
	VcsId   models.Id `json:"vcsId" binding:"required,startswith=vcs-,max=32" example:"vcs-c3lcrjxczjdywmk0go90"`
	RepoId  string    `json:"repoId" binding:"required,max=128" example:"1234567890"`
	GitTags string    `json:"gitTags" example:"Git Tags" binding:"max=128"`
	Branch  string    `json:"branch" example:"master" binding:"max=128"`
	Dir     string    `json:"dir" example:"/" binding:"max=255"`
}

type SearchPolicyGroupForm struct {
	NoPageSizeForm

	Q string `form:"q" json:"q" binding:""` // 策略组名称，支持模糊搜索
}

type UpdatePolicyGroupForm struct {
	BaseForm
	Id          models.Id `uri:"id" binding:"required,startswith=pog-,max=32" swaggerignore:"true"`
	Name        string    `json:"name" form:"name" binding:"omitempty,gte=2,lte=64"`
	Description string    `json:"description" binding:"max=255" example:"本组包含对于安全合规的检查策略"`
	Enabled     bool      `json:"enabled" form:"enabled" binding:""`
	Labels      []string  `json:"labels" binding:"omitempty,dive,required,max=128" example:"[security,alicloud]"`
	Source      string    `json:"source" binding:"omitempty,oneof=vcs registry" enums:"vcs,registry" example:"来源"`
	VcsId       models.Id `json:"vcsId" binding:"omitempty,startswith=vcs-,max=32" example:"vcs-c3lcrjxczjdywmk0go90"`
	RepoId      string    `json:"repoId" binding:"max=128" example:"1234567890"`
	GitTags     string    `json:"gitTags" example:"Git Tags" binding:"max=128"`
	Branch      string    `json:"branch" example:"master" binding:"max=128"`
	Dir         string    `json:"dir" example:"/" binding:"max=255"`
}

type DeletePolicyGroupForm struct {
	BaseForm

	Id models.Id `uri:"id" binding:"required,startswith=pog-,max=32"`
}

type DetailPolicyGroupForm struct {
	BaseForm

	Id models.Id `uri:"id" binding:"required,startswith=pog-,max=32"`
}

type UpdatePolicyRelForm struct {
	BaseForm
	Id             models.Id   `uri:"id" binding:"required,max=32" example:"tpl-c3ek0co6n88ldvq1n6ag" swaggerignore:"true"`
	PolicyGroupIds []models.Id `json:"policyGroupIds" binding:"required,dive,required,startswith=pog-,max=32" example:"pog-c3ek0co6n88ldvq1n6ag,pog-c3ek0co6n88ldvq1n6bg"`
	Scope          string      `json:"-" swaggerignore:"true" binding:"omitempty,oneof=env template"`
}

type EnableScanForm struct {
	BaseForm

	Id      models.Id `uri:"id" swaggerignore:"true" binding:"omitempty,startswith=env-,max=32" example:"env-c3ek0co6n88ldvq1n6ag"` // ID
	Scope   string    `json:"-" swaggerignore:"true" binding:"omitempty,oneof=env template"`
	Enabled bool      `json:"enabled" binding:"" example:"true"` // 是否启用扫描
}

type ScanTemplateForm struct {
	BaseForm

	Id    models.Id `uri:"id" binding:"required,startswith=tpl-,max=32" example:"tpl-c3ek0co6n88ldvq1n6ag"` // 云模板Id
	Parse bool      `json:"parse" binding:""  enums:"true,false" example:"false"`                           // 是否只执行解析
}

type ScanTemplateForms struct {
	BaseForm

	Ids   []models.Id `json:"ids" binding:"required,dive,required,startswith=tpl-,max=32" example:"[tpl-c3ek0co6n88ldvq1n6ag, tpl-c3ek0co6n88ldvasdn6ag]"` // 云模板Id
	Parse bool        `json:"parse" binding:""  enums:"true,false" example:"false"`                                                                        // 是否只执行解析
}

type ScanEnvironmentForm struct {
	BaseForm
	Id    models.Id `uri:"id" binding:"required,startswith=env-,max=32" example:"env-c3ek0co6n88ldvq1n6ag" swaggerignore:"true"` // 环境Id
	Parse bool      `json:"parse" binding:""  enums:"true,false" example:"false"`                                                // 是否只执行解析
}

type PolicyParseForm struct {
	BaseForm

	TemplateId models.Id `form:"tplId" json:"tplId" binding:"required,startswith=tpl-,max=32" example:"tpl-c3ek0co6n88ldvq1n6ag"`  // 云模板Id
	EnvId      models.Id `form:"envId" json:"envId" binding:"omitempty,startswith=env-,max=32" example:"env-c3ek0co6n88ldvq1n6ag"` // 环境Id
}

type OpnPolicyAndPolicyGroupRelForm struct {
	BaseForm
	PolicyGroupId models.Id `uri:"id" json:"policyGroupId" form:"policyGroupId" binding:"required,startswith=pog-,max=32" swaggerignore:"true"`
	RmPolicyIds   []string  `json:"rmPolicyIds" binding:"omitempty,dive,required,startswith=po-,max=32" example:"po-c3ek0co6n88ldvq1n6ag"`
	AddPolicyIds  []string  `json:"addPolicyIds" binding:"omitempty,dive,required,startswith=po-,max=32" example:"po-c3ek0co6n88ldvq1n6ag"`
}

type CreatePolicySuppressForm struct {
	BaseForm

	CreatorId models.Id   `json:"creatorId" `
	TplId     models.Id   `json:"tplId" binding:""`
	EnvId     []models.Id `json:"envId" `
	PolicyId  models.Id   `json:"policyId" binding:""`

	Reason string `json:"reason" form:"reason" `
	Type   string `json:"type" form:"type" enums:"'strategy','source'"`
}

type SearchPolicySuppressForm struct {
	PageForm
	Id models.Id `uri:"id" binding:"required,startswith=po-,max=32" swaggerignore:"true"`
}

type SearchPolicySuppressSourceForm struct {
	NoPageSizeForm
	Id models.Id `uri:"id" binding:"required,startswith=po-,max=32" swaggerignore:"true"`
}

type DeletePolicySuppressForm struct {
	BaseForm

	Id         models.Id `uri:"id" binding:"required,startswith=po-,max=32"`
	SuppressId models.Id `uri:"suppressId" binding:"required"`
}

type SearchPolicyTplForm struct {
	NoPageSizeForm

	TplId models.Id `form:"tplId" binding:"omitempty,startswith=tpl-,max=32"`
	Q     string    `form:"q" json:"q" binding:""` // 模糊搜索
}

type DetailPolicyTplForm struct {
	BaseForm
	Id models.Id `json:"id" form:"id" binding:"required,max=32"`
}

type TplOfPolicyForm struct {
	PageForm

	Id       models.Id `json:"id" form:"id" binding:"required,startswith=tpl-,max=32"`
	Q        string    `form:"q" json:"q" binding:""` // 策略组名称，支持模糊搜索
	Severity string    `json:"severity" form:"severity" binding:"omitempty,oneof=high medium low none" enums:"'high','medium','low','none'" example:"medium"`
	GroupId  models.Id `json:"groupId" form:"groupId" binding:"omitempty,startswith=pog-,max=32"`
}

type TplOfPolicyGroupForm struct {
	NoPageSizeForm

	Id models.Id `uri:"id" binding:"required,startswith=tpl-,max=32"`
}

type SearchPolicyEnvForm struct {
	NoPageSizeForm

	ProjectId models.Id `form:"projectId" binding:"omitempty,startswith=p-,max=32"`
	EnvId     models.Id `form:"envId" binding:"omitempty,startswith=env-,max=32"`
	Q         string    `form:"q" json:"q" binding:""` // 模糊搜索
}

type EnvOfPolicyForm struct {
	PageForm
	Id       models.Id `json:"id" form:"id" binding:"omitempty,startswith=env-,max=32" swaggerignore:"true"`
	Q        string    `form:"q" json:"q" binding:""` // 策略组名称，支持模糊搜索
	Severity string    `json:"severity" form:"severity" binding:"omitempty,oneof=high medium low none" enums:"'high','medium','low','none'" example:"medium"`
	GroupId  models.Id `json:"groupId" form:"groupId" binding:"omitempty,startswith=pog-,max=32"`
}

type PolicyErrorForm struct {
	PageForm
	Id models.Id `uri:"id" binding:"required,startswith=po-,max=32" swaggerignore:"true"`
	Q  string    `json:"q" form:"q"`
}

type UpdatePolicySuppressForm struct {
	BaseForm
	Id           models.Id   `uri:"id" swaggerignore:"true" binding:"required,startswith=po-,max=32"`                                       // 策略ID
	Reason       string      `json:"reason" example:"测试环境无需检测" binding:"omitempty,max=255"`                                                 // 屏蔽原因
	AddSourceIds []models.Id `json:"addTargetIds" example:"env-c3ek0co6n88ldvq1n6ag" binding:"required,dive,required,startswith=r-,max=32"` // 添加屏蔽源ID列表
	//RmSourceIds  []models.Id `json:"rmTargetIds" example:"env-c3ek0co6n88ldvq1n6ag"`  // 删除屏蔽源ID列表
}

type PolicyScanResultForm struct {
	NoPageSizeForm
	Id     models.Id `uri:"id" binding:"required,startswith=env-,max=32" swaggerignore:"true"`                                   // 环境ID
	TaskId models.Id `json:"taskId" form:"taskId" binding:"omitempty,startswith=run-,max=32" example:"run-c3ek0co6n88ldvq1n6ag"` // 任务ID
}

type PolicyScanReportForm struct {
	BaseForm

	Id        models.Id `uri:"id" binding:"required,max=32" swaggerignore:"true"`     // 策略/策略组ID
	From      time.Time `json:"from" form:"from" example:"2006-01-02T15:04:05Z07:00"` //  开始日期
	To        time.Time `json:"to" form:"to" example:"2006-01-02T15:04:05Z07:00"`     // 结束日期
	ShowCount int       `json:"showCount" form:"showCount" example:"5"`
}

type PolicyTestForm struct {
	BaseForm

	Input string `form:"input" json:"input" binding:"required" example:"{\n\"alicloud_instance\": [\n\n{\t\n\"id\": \"alicloud_instance.instance\"..."` // 脚本验证源数据
	Rego  string `form:"rego" json:"rego" binding:"required" example:"package accurics\ninstanceWithNoVpc[retVal] {..."`                                // rego脚本内容
}

type PolicyLastTasksForm struct {
	PageForm
	Id    models.Id `uri:"id" binding:"required,startswith=pog-,max=32" swaggerignore:"true"`
	Scope string    `json:"-"`
}

type SearchGroupOfPolicyForm struct {
	PageForm

	Id     models.Id `uri:"id" binding:"required,startswith=pog-,max=32"`
	IsBind bool      `json:"bind" form:"bind" binding:""` //  ture: 查询绑定策略组的策略，false: 查询未绑定的策略组的策略
}

type PolicyGroupChecksForm struct {
	BaseForm
	Name         string    `json:"name" form:"name"`
	RepoId       string    `json:"repoId" form:"repoId" binding:"required"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision" binding:"required"`
	VcsId        models.Id `json:"vcsId" form:"vcsId" binding:"required,startswith=vcs-,max=32"`
	Dir          string    `json:"dir" form:"dir"`
	TemplateId   models.Id `json:"templateId" form:"templateId"`
}
