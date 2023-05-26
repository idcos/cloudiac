// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package forms

import (
	"cloudiac/portal/models"
	"time"
)

type envDestroyTtlForm struct {
	TTL       string `form:"ttl" json:"ttl" binding:"omitempty,oneof=0 12h 1d 3d 1w 15d 30d" enums:"0,12h,1d,3d,1w,15d,30d"` // 存活时间
	DestroyAt string `form:"destroyAt" json:"destroyAt" binding:""`                                                          // 自动销毁时间(时间戳)
}

type envDeployTtlForm struct {
	DeployAt string `form:"deployAt" json:"deployAt" binding:""` // 自动部署时间(时间戳)
}

type CreateEnvForm struct {
	BaseForm
	envDestroyTtlForm
	envDeployTtlForm

	TplId    models.Id `form:"tplId" json:"tplId" binding:"required,startswith=tpl-,max=32"`                 // 模板ID
	Name     string    `form:"name" json:"name" binding:"required,gte=2,lte=64"`                             // 环境名称
	OneTime  bool      `form:"oneTime" json:"oneTime" binding:""`                                            // 一次性环境标识
	Triggers []string  `form:"triggers" json:"triggers" binding:"omitempty,dive,required,oneof=commit prmr"` // 启用触发器，触发器：commit（每次推送自动部署），prmr（提交PR/MR的时候自动执行plan）

	Tags string `form:"tags" json:"tags" binding:"max=255"` // 环境的 tags，多个 tag 以 "," 分隔

	AutoApproval    bool       `form:"autoApproval" json:"autoApproval"  binding:"" enums:"true,false"` // 是否自动审批
	StopOnViolation bool       `form:"stopOnViolation" json:"stopOnViolation" enums:"true,false"`       // 合规不通过是否中止任务
	TaskType        string     `form:"taskType" json:"taskType" binding:"required" enums:"plan,apply"`  // 环境创建后触发的任务步骤，plan计划,apply部署
	Targets         string     `form:"targets" json:"targets" binding:""`                               // Terraform target 参数列表，多个参数用 , 进行分隔
	RunnerId        string     `form:"runnerId" json:"runnerId" binding:"max=32"`                       // 环境默认部署通道
	RunnerTags      []string   `form:"runnerTags" json:"runnerTags" binding:"omitempty,dive,max=256"`   // 环境默认部署通道tags
	Revision        string     `form:"revision" json:"revision" binding:"max=64"`                       // 分支/标签
	StepTimeout     int        `form:"stepTimeout" json:"stepTimeout" binding:""`                       // 部署超时时间（单位：秒）
	Variables       []Variable `form:"variables" json:"variables" binding:"omitempty,dive,required"`    // 自定义变量列表，该变量列表会覆盖现有的变量

	TfVarsFile   string    `form:"tfVarsFile" json:"tfVarsFile" binding:"max=255"`              // Terraform tfvars 变量文件路径
	PlayVarsFile string    `form:"playVarsFile" json:"playVarsFile" binding:"max=255"`          // Ansible playbook 变量文件路径
	Playbook     string    `form:"playbook" json:"playbook" binding:"omitempty,max=255"`        // Ansible playbook 入口文件路径
	KeyId        models.Id `form:"keyId" json:"keyId" binding:"omitempty,startswith=k-,max=32"` // 部署密钥ID
	KeyName      string    `form:"keyName" json:"keyName" binding:"omitempty,max=255"`          // 部署密钥名称
	Workdir      string    `form:"workdir" json:"workdir" `                                     // 工作目录

	RetryNumber int         `form:"retryNumber" json:"retryNumber" binding:""` // 重试总次数
	RetryDelay  int         `form:"retryDelay" json:"retryDelay" binding:""`   // 重试时间间隔
	RetryAble   bool        `form:"retryAble" json:"retryAble" binding:""`     // 是否允许任务进行重试
	ExtraData   models.JSON `form:"extraData" json:"extraData" binding:""`     // 扩展字段，用于存储外部服务调用时的信息

	VarGroupIds    []models.Id `json:"varGroupIds" form:"varGroupIds" binding:"omitempty,dive,required,startswith=vg-,max=32"`
	DelVarGroupIds []models.Id `json:"delVarGroupIds" form:"delVarGroupIds" binding:"omitempty,dive,required,startswith=vg-,max=32"`

	SampleVariables []SampleVariables `json:"sampleVariables" form:"sampleVariables" binding:"omitempty,dive,required"`

	Callback string `json:"callback" form:"callback" binding:"max=255"` // 外部请求的回调方式

	CronDriftExpress string `json:"cronDriftExpress" form:"cronDriftExpress" binding:"max=255"` // 偏移检测表达式
	AutoRepairDrift  bool   `json:"autoRepairDrift" form:"autoRepairDrift"`                     // 是否进行自动纠偏
	OpenCronDrift    bool   `json:"openCronDrift" form:"openCronDrift" binding:""`              // 是否开启偏移检测

	PolicyEnable bool        `json:"policyEnable" form:"policyEnable" binding:""`                                             // 是否开启合规检测
	PolicyGroup  []models.Id `json:"policyGroup" form:"policyGroup" binding:"omitempty,dive,required,startswith=pog-,max=32"` // 绑定策略组集合

	Source string `json:"source" form:"source" binding:"required"` // 调用来源

	AutoDeployCron  string `json:"autoDeployCron" form:"autoDeployCron"`   // 自动部署任务的Cron表达式
	AutoDestroyCron string `json:"autoDestroyCron" form:"autoDestroyCron"` // 自动销毁任务的Cron表达式

	EnvTags  []models.Tag `json:"envTags" form:"envTags" `   // 外部系统传入的 tags
	UserTags []models.Tag `json:"userTags" form:"userTags" ` // 用户设置的 tags
}

type SampleVariables struct {
	Name      string `json:"name" form:"name" binding:"required,lte=64"`
	Value     string `json:"value" form:"value" binding:""`
	Sensitive bool   `json:"sensitive" form:"sensitive" binding:""`
}

type CronDriftForm struct {
	BaseForm
	CronDriftExpress string `json:"cronDriftExpress" form:"cronDriftExpress" binding:"max=255"` // 偏移检测表达式
	AutoRepairDrift  bool   `json:"autoRepairDrift" form:"autoRepairDrift"`                     // 是否进行自动纠偏
	OpenCronDrift    bool   `json:"openCronDrift" form:"openCronDrift"`                         // 是否开启偏移检测
}

type UpdateEnvForm struct {
	BaseForm
	envDestroyTtlForm
	envDeployTtlForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=env-,max=32"` // 环境ID，swagger 参数通过 param path 指定，这里忽略

	Name        string    `form:"name" json:"name" binding:"omitempty,gte=2,lte=64"`           // 环境名称
	Description string    `form:"description" json:"description" binding:"max=255"`            // 环境描述
	KeyId       models.Id `form:"keyId" json:"keyId" binding:"omitempty,startswith=k-,max=32"` // 部署密钥ID
	KeyName     string    `form:"keyName" json:"keyName" binding:"omitempty,max=255"`          // 部署密钥名称
	RunnerId    string    `form:"runnerId" json:"runnerId" binding:"max=32"`                   // 环境默认部署通道
	Archived    bool      `form:"archived" json:"archived" enums:"true,false"`                 // 归档状态，默认返回未归档环境
	Tags        string    `form:"tags" json:"tags" binding:"max=255"`                          // 环境的 tags，多个 tag 以 "," 分隔
	StepTimeout int       `form:"stepTimeout" json:"stepTimeout" binding:""`                   // 部署超时时间（单位：秒）

	AutoApproval    bool `form:"autoApproval" json:"autoApproval"  binding:"" enums:"true,false"` // 是否自动审批
	StopOnViolation bool `form:"stopOnViolation" json:"stopOnViolation" enums:"true,false"`       // 合规不通过是否中止任务

	Triggers         []string `form:"triggers" json:"triggers" binding:"omitempty,dive,required,oneof=commit prmr"` // 启用触发器，触发器：commit（每次推送自动部署），prmr（提交PR/MR的时候自动执行plan）
	RetryNumber      int      `form:"retryNumber" json:"retryNumber" binding:""`                                    // 重试总次数
	RetryDelay       int      `form:"retryDelay" json:"retryDelay" binding:""`                                      // 重试时间间隔
	RetryAble        bool     `form:"retryAble" json:"retryAble" binding:""`                                        // 是否允许任务进行重试
	CronDriftExpress string   `json:"cronDriftExpress" form:"cronDriftExpress" binding:"max=255"`                   // 偏移检测表达式
	AutoRepairDrift  bool     `json:"autoRepairDrift" form:"autoRepairDrift"`                                       // 是否进行自动纠偏
	OpenCronDrift    bool     `json:"openCronDrift" form:"openCronDrift"`                                           // 是否开启偏移检测

	PolicyEnable bool        `json:"policyEnable" form:"policyEnable"`                                                        // 是否开启合规检测
	PolicyGroup  []models.Id `json:"policyGroup" form:"policyGroup" binding:"omitempty,dive,required,startswith=pog-,max=32"` // 绑定策略组集合

	AutoDeployCron  string `json:"autoDeployCron" form:"autoDeployCron"`   // 自动部署任务的Cron表达式
	AutoDestroyCron string `json:"autoDestroyCron" form:"autoDestroyCron"` // 自动销毁任务的Cron表达式
}

type DeployEnvForm struct {
	BaseForm
	envDestroyTtlForm
	envDeployTtlForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=env-,max=32"` // 环境ID，swagger 参数通过 param path 指定，这里忽略

	Name            string   `form:"name" json:"name" binding:"omitempty,gte=2,lte=64"`                            // 环境名称
	Triggers        []string `form:"triggers" json:"triggers" binding:"omitempty,dive,required,oneof=commit prmr"` // 启用触发器，触发器：commit（每次推送自动部署），prmr（提交PR/MR的时候自动执行plan）
	AutoApproval    bool     `form:"autoApproval" json:"autoApproval"  binding:"" enums:"true,false"`              // 是否自动审批
	StopOnViolation bool     `form:"stopOnViolation" json:"stopOnViolation" enums:"true,false"`                    // 合规不通过是否中止任务

	TaskType    string   `form:"taskType" json:"taskType" binding:"required,oneof=plan apply destroy" enums:"plan,apply,destroy"` // 环境创建后触发的任务步骤，plan计划,apply部署,destroy销毁资源
	Targets     string   `form:"targets" json:"targets" binding:""`                                                               // Terraform target 参数列表
	RunnerId    string   `form:"runnerId" json:"runnerId" binding:"max=32"`                                                       // 环境默认部署通道
	RunnerTags  []string `form:"runnerTags" json:"runnerTags" binding:"omitempty,dive,required,max=256"`                          // 环境默认部署通道Tags
	Revision    string   `form:"revision" json:"revision" binding:"max=64"`                                                       // 分支/标签
	StepTimeout int      `form:"stepTimeout" json:"stepTimeout" binding:""`                                                       // 部署超时时间（单位：秒）

	RetryNumber int  `form:"retryNumber" json:"retryNumber" binding:""` // 重试总次数
	RetryDelay  int  `form:"retryDelay" json:"retryDelay" binding:""`   // 重试时间间隔
	RetryAble   bool `form:"retryAble" json:"retryAble" binding:""`     // 是否允许任务进行重试

	ExtraData models.JSON `form:"extraData" json:"extraData" binding:""` // 扩展字段，用于存储外部服务调用时的信息

	Variables []Variable `form:"variables" json:"variables" binding:"omitempty,dive,required"` // 自定义变量列表，该变量列表会覆盖现有的变量

	TfVarsFile   string    `form:"tfVarsFile" json:"tfVarsFile" binding:"max=255"`              // Terraform tfvars 变量文件路径
	PlayVarsFile string    `form:"playVarsFile" json:"playVarsFile" binding:"max=255"`          // Ansible playbook 变量文件路径
	Playbook     string    `form:"playbook" json:"playbook" binding:"omitempty,max=255"`        // Ansible playbook 入口文件路径
	KeyId        models.Id `form:"keyId" json:"keyId" binding:"omitempty,startswith=k-,max=32"` // 部署密钥ID
	KeyName      string    `form:"keyName" json:"keyName" binding:"omitempty,max=255"`          // 部署密钥名称
	Workdir      string    `form:"workdir" json:"workdir" binding:"max=32"`                     // 工作目录

	VarGroupIds    []models.Id `json:"varGroupIds" form:"varGroupIds" binding:"omitempty,dive,required,startswith=vg-,max=32"`
	DelVarGroupIds []models.Id `json:"delVarGroupIds" form:"delVarGroupIds" binding:"omitempty,dive,required,startswith=vg-,max=32"`

	CronDriftExpress string `json:"cronDriftExpress" form:"cronDriftExpress" binding:"max=255"` // 偏移检测表达式
	AutoRepairDrift  bool   `json:"autoRepairDrift" form:"autoRepairDrift"`                     // 是否进行自动纠偏
	OpenCronDrift    bool   `json:"openCronDrift" form:"openCronDrift" binding:""`              // 是否开启偏移检测

	PolicyEnable bool        `json:"policyEnable" form:"policyEnable" binding:""`                                             // 是否开启合规检测
	PolicyGroup  []models.Id `json:"policyGroup" form:"policyGroup" binding:"omitempty,dive,required,startswith=pog-,max=32"` // 绑定策略组集合

	Source string `json:"source" form:"source" binding:"required"` // 调用来源

	AutoDeployCron  string `json:"autoDeployCron" form:"autoDeployCron"`   // 自动部署任务的Cron表达式
	AutoDestroyCron string `json:"autoDestroyCron" form:"autoDestroyCron"` // 自动销毁任务的Cron表达式

	// 部署plan任务时生效，进行漂移检测时，从最后一次任务获取配置信息进行检测
	IsDriftTask bool `json:"isDriftTask" form:"isDriftTask" `

	EnvTags  []models.Tag `json:"envTags" form:"envTags" `
	UserTags []models.Tag `json:"userTags" form:"userTags" `
}

type ArchiveEnvForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=env-,max=32"` // 环境ID，swagger 参数通过 param path 指定，这里忽略

	Archived bool `form:"archived" json:"archived" binding:"required" enums:"true,false"` // 归档状态
}

type SearchEnvForm struct {
	NoPageSizeForm

	Q        string `form:"q" json:"q" binding:""`                                                                    // 环境名称，支持模糊查询
	Status   string `form:"status" json:"status" enums:"active,failed,inactive,running,approving"`                    // 环境状态，active活跃, inactive非活跃,failed错误,running部署中,approving审批中
	Archived string `form:"archived" json:"archived" binding:"omitempty,oneof=true false all" enums:"true,false,all"` // 归档状态，默认返回未归档环境

	Deploying *bool `form:"deploying" json:"deploying" binding:""` // 环境部署状态，不传则表示不过滤部署状态

	StartTime *time.Time `json:"startTime" form:"startTime" `
	EndTime   *time.Time `json:"endTime" form:"endTime" `
}

type DeleteEnvForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=env-,max=32"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type DetailEnvForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=env-,max=32"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type EnvParam struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=env-,max=32"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type SearchEnvResourceForm struct {
	NoPageSizeForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=env-,max=32"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
	Q  string    `form:"q" json:"q" binding:""`                                                      // 资源名称，支持模糊查询
}

type DestroyEnvForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=env-,max=32"` // 环境ID，swagger 参数通过 param path 指定，这里忽略

	Source string `json:"source" form:"source" binding:"required"` // 调用来源
}

type SearchEnvVariableForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=env-,max=32"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type SearchEnvResourceGraphForm struct {
	BaseForm

	Id        models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=env-,max=32"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
	Dimension string    `json:"dimension" form:"dimension" binding:"required"`                              // 资源名称，支持模糊查询
}

type ResourceGraphDetailForm struct {
	BaseForm

	Id         models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=env-,max=32"`             // 环境ID，swagger 参数通过 param path 指定，这里忽略
	ResourceId models.Id `uri:"resourceId" json:"resourceId" swaggerignore:"true" binding:"required,contains=r-,max=32"` // 部署成功后后资源ID
}

type UpdateEnvTagsForm struct {
	BaseForm

	Id   models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=env-,max=32"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
	Tags string    `json:"tags" form:"tags" binding:"max=255"`
}

type EnvLockForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type EnvUnLockForm struct {
	BaseForm

	Id             models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
	ClearDestroyAt bool      `json:"clearDestroyAt" form:"clearDestroyAt" `
}

type EnvUnLockConfirmForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}
