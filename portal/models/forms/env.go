// Copyright 2021 CloudJ Company Limited. All rights reserved.

package forms

import (
	"cloudiac/portal/models"
)

type envTtlForm struct {
	TTL       string `form:"ttl" json:"ttl" binding:"" enums:"0,12h,1d,3d,1w,15d,30d"` // 存活时间
	DestroyAt string `form:"destroyAt" json:"destroyAt" binding:""`                    // 自动销毁时间(时间戳)
}

type CreateEnvForm struct {
	BaseForm
	envTtlForm

	TplId    models.Id `form:"tplId" json:"tplId" binding:"required"`            // 模板ID
	Name     string    `form:"name" json:"name" binding:"required,gte=2,lte=64"` // 环境名称
	OneTime  bool      `form:"oneTime" json:"oneTime" binding:""`                // 一次性环境标识
	Triggers []string  `form:"triggers" json:"triggers" binding:""`              // 启用触发器，触发器：commit（每次推送自动部署），prmr（提交PR/MR的时候自动执行plan）

	AutoApproval    bool   `form:"autoApproval" json:"autoApproval"  binding:"" enums:"true,false"` // 是否自动审批
	StopOnViolation bool   `form:"stopOnViolation" json:"stopOnViolation" enums:"true,false"`       // 合规不通过是否中止任务
	TaskType        string `form:"taskType" json:"taskType" binding:"required" enums:"plan,apply"`  // 环境创建后触发的任务步骤，plan计划,apply部署
	Targets         string `form:"targets" json:"targets" binding:""`                               // Terraform target 参数列表，多个参数用 , 进行分隔
	RunnerId        string `form:"runnerId" json:"runnerId" binding:""`                             // 环境默认部署通道
	Revision        string `form:"revision" json:"revision" binding:""`                             // 分支/标签
	Timeout         int    `form:"timeout" json:"timeout" binding:""`                               // 部署超时时间（单位：秒）

	Variables []Variables `form:"variables" json:"variables" binding:""` // 自定义变量列表，该变量列表会覆盖现有的变量

	TfVarsFile   string    `form:"tfVarsFile" json:"tfVarsFile" binding:""`     // Terraform tfvars 变量文件路径
	PlayVarsFile string    `form:"playVarsFile" json:"playVarsFile" binding:""` // Ansible playbook 变量文件路径
	Playbook     string    `form:"playbook" json:"playbook" binding:""`         // Ansible playbook 入口文件路径
	KeyId        models.Id `form:"keyId" json:"keyId" binding:""`               // 部署密钥ID

	RetryNumber int         `form:"retryNumber" json:"retryNumber" binding:""` // 重试总次数
	RetryDelay  int         `form:"retryDelay" json:"retryDelay" binding:""`   // 重试时间间隔
	RetryAble   bool        `form:"retryAble" json:"retryAble" binding:""`     // 是否允许任务进行重试
	ExtraData   models.JSON `form:"extraData" json:"extraData" binding:""`     // 扩展字段，用于存储外部服务调用时的信息

	VarGroupIds    []models.Id `json:"varGroupIds" form:"varGroupIds" `
	DelVarGroupIds []models.Id `json:"delVarGroupIds" form:"delVarGroupIds" `

	SampleVariables []SampleVariables `json:"sampleVariables" form:"sampleVariables" `

	Callback string `json:"callback" form:"callback"` // 外部请求的回调方式
}

type SampleVariables struct {
	Name  string `json:"name" form:"name" `
	Value string `json:"value" form:"value" `
}

type UpdateEnvForm struct {
	BaseForm
	envTtlForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略

	Name        string    `form:"name" json:"name" binding:""`                      // 环境名称
	Description string    `form:"description" json:"description" binding:"max=255"` // 环境描述
	KeyId       models.Id `form:"keyId" json:"keyId" binding:""`                    // 部署密钥ID
	RunnerId    string    `form:"runnerId" json:"runnerId" binding:""`              // 环境默认部署通道
	Archived    bool      `form:"archived" json:"archived" enums:"true,false"`      // 归档状态，默认返回未归档环境

	AutoApproval    bool `form:"autoApproval" json:"autoApproval"  binding:"" enums:"true,false"` // 是否自动审批
	StopOnViolation bool `form:"stopOnViolation" json:"stopOnViolation" enums:"true,false"`       // 合规不通过是否中止任务

	Triggers    []string `form:"triggers" json:"triggers" binding:""`       // 启用触发器，触发器：commit（每次推送自动部署），prmr（提交PR/MR的时候自动执行plan）
	RetryNumber int      `form:"retryNumber" json:"retryNumber" binding:""` // 重试总次数
	RetryDelay  int      `form:"retryDelay" json:"retryDelay" binding:""`   // 重试时间间隔
	RetryAble   bool     `form:"retryAble" json:"retryAble" binding:""`     // 是否允许任务进行重试
}

type DeployEnvForm struct {
	BaseForm
	envTtlForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略

	Name            string   `form:"name" json:"name" binding:""`                                     // 环境名称
	Triggers        []string `form:"triggers" json:"triggers" binding:""`                             // 启用触发器，触发器：commit（每次推送自动部署），prmr（提交PR/MR的时候自动执行plan）
	AutoApproval    bool     `form:"autoApproval" json:"autoApproval"  binding:"" enums:"true,false"` // 是否自动审批
	StopOnViolation bool     `form:"stopOnViolation" json:"stopOnViolation" enums:"true,false"`       // 合规不通过是否中止任务

	TaskType string `form:"taskType" json:"taskType" binding:"required" enums:"plan,apply,destroy"` // 环境创建后触发的任务步骤，plan计划,apply部署,destroy销毁资源
	Targets  string `form:"targets" json:"targets" binding:""`                                      // Terraform target 参数列表
	RunnerId string `form:"runnerId" json:"runnerId" binding:""`                                    // 环境默认部署通道
	Revision string `form:"revision" json:"revision" binding:""`                                    // 分支/标签
	Timeout  int    `form:"timeout" json:"timeout" binding:""`                                      // 部署超时时间（单位：秒）

	RetryNumber int  `form:"retryNumber" json:"retryNumber" binding:""` // 重试总次数
	RetryDelay  int  `form:"retryDelay" json:"retryDelay" binding:""`   // 重试时间间隔
	RetryAble   bool `form:"retryAble" json:"retryAble" binding:""`     // 是否允许任务进行重试

	Variables         []Variables `form:"variables" json:"variables" binding:""`       // 自定义变量列表，该变量列表会覆盖现有的变量
	DeleteVariablesId []string    `json:"deleteVariablesId" form:"deleteVariablesId" ` //删除的变量id

	TfVarsFile   string    `form:"tfVarsFile" json:"tfVarsFile" binding:""`     // Terraform tfvars 变量文件路径
	PlayVarsFile string    `form:"playVarsFile" json:"playVarsFile" binding:""` // Ansible playbook 变量文件路径
	Playbook     string    `form:"playbook" json:"playbook" binding:""`         // Ansible playbook 入口文件路径
	KeyId        models.Id `form:"keyId" json:"keyId" binding:""`               // 部署密钥ID

	VarGroupIds    []models.Id `json:"varGroupIds" form:"varGroupIds" `
	DelVarGroupIds []models.Id `json:"delVarGroupIds" form:"delVarGroupIds" `
}

type ArchiveEnvForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略

	Archived bool `form:"archived" json:"archived" binding:"required" enums:"true,false"` // 归档状态
}

type SearchEnvForm struct {
	PageForm

	Q        string `form:"q" json:"q" binding:""`                                                 // 环境名称，支持模糊查询
	Status   string `form:"status" json:"status" enums:"active,failed,inactive,running,approving"` // 环境状态，active活跃, inactive非活跃,failed错误,running部署中,approving审批中
	Archived string `form:"archived" json:"archived" enums:"true,false,all"`                       // 归档状态，默认返回未归档环境
}

type DeleteEnvForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type DetailEnvForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type EnvParam struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type SearchEnvResourceForm struct {
	PageForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
	Q  string    `form:"q" json:"q" binding:""`            // 资源名称，支持模糊查询
}

type DestroyEnvForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type SearchEnvVariableForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type SearchEnvResourceGraphForm struct {
	BaseForm

	Id        models.Id `uri:"id" json:"id" swaggerignore:"true"`              // 环境ID，swagger 参数通过 param path 指定，这里忽略
	Dimension string    `json:"dimension" form:"dimension" binding:"required"` // 资源名称，支持模糊查询
}

type ResourceGraphDetailForm struct {
	BaseForm

	Id         models.Id `uri:"id" json:"id" swaggerignore:"true"`                 // 环境ID，swagger 参数通过 param path 指定，这里忽略
	ResourceId models.Id `uri:"resourceId" json:"resourceId" swaggerignore:"true"` // 部署成功后后资源ID
}
