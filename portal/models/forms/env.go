package forms

import "cloudiac/portal/models"

type CreateEnvForm struct {
	BaseForm
	TplId        models.Id `form:"tplId" json:"tplId" binding:"required"`                           // 模板ID
	Name         string    `form:"name" json:"name" binding:"required,gte=2,lte=32"`                // 环境名称
	OneTime      bool      `form:"oneTime" json:"oneTime" binding:""`                               // 一次性环境标识
	Triggers     []string  `form:"triggers" json:"triggers" binding:""`                             // 启用触发器，触发器：commit（每次推送自动部署），prmr（提交PR/MR的时候自动执行plan）
	AutoApproval bool      `form:"autoApproval" json:"autoApproval"  binding:"" enums:"true,false"` // 是否自动审批
	DestroyAt    string    `form:"destroyAt" json:"destroyAt" binding:"" `                          // 自动销毁时间， 0: 不自动销毁, 12h: 12小时,	1d：一天,3d: 三天	1w: 一周，7天	15d: 半个月，15天	1m: 一个月，28/29/30/31天根据不同月份	xxxx-xx-xx xx:xx：指定时间（选择指定时间时出现时间选择框），格式：年-月-日 时:分

	TaskType string `form:"taskType" json:"taskType" binding:"required" enums:"plan,apply"` // 环境创建后触发的任务步骤，plan计划,apply部署
	Targets  string `form:"targets" json:"targets" binding:""`                              // Terraform target 参数列表，多个参数用 , 进行分隔
	RunnerId string `form:"runnerId" json:"runnerId" binding:""`                            // 环境默认部署通道
	Revision string `form:"revision" json:"revision" binding:""`                            // 分支/标签
	Timeout  int    `form:"timeout" json:"timeout" binding:""`                              // 部署超时时间（单位：秒）

	Variables []models.VariableBody `form:"variables" json:"variable" binding:""` // 自定义变量列表，该变量列表会覆盖现有的变量

	TfVarsFile   string `form:"tfVarsFile" json:"tfVarsFile" binding:""`     // Terraform tfvars 变量文件路径
	PlayVarsFile string `form:"playVarsFile" json:"playVarsFile" binding:""` // Ansible playbook 变量文件路径
	Playbook     string `form:"playbook" json:"playbook" binding:""`         // Ansible playbook 入口文件路径
}

type UpdateEnvForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略

	Name          string `form:"name" json:"name" binding:""`                                     // 环境名称
	Description   string `form:"description" json:"description" binding:"max=255"`                // 环境描述
	RunnerId      string `form:"runnerId" json:"runnerId" binding:""`                             // 环境默认部署通道
	Archived      bool   `form:"archived" json:"archived" enums:"true,false"`                     // 归档状态，默认返回未归档环境
	AutoApproval  bool   `form:"autoApproval" json:"autoApproval"  binding:"" enums:"true,false"` // 是否自动审批
	AutoDestroyAt string `form:"destroyAt" json:"destroyAt" binding:"" `                          // 自动销毁时间， 0: 不自动销毁, 12h: 12小时,	1d：一天,3d: 三天	1w: 一周，7天	15d: 半个月，15天	1m: 一个月，28/29/30/31天根据不同月份	xxxx-xx-xx xx:xx：指定时间（选择指定时间时出现时间选择框），格式：年-月-日 时:分
}

type DeployEnvForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略

	Name          string   `form:"name" json:"name" binding:""`                                     // 环境名称
	Triggers      []string `form:"triggers" json:"triggers" binding:""`                             // 启用触发器，触发器：commit（每次推送自动部署），prmr（提交PR/MR的时候自动执行plan）
	AutoApproval  bool     `form:"autoApproval" json:"autoApproval"  binding:"" enums:"true,false"` // 是否自动审批
	AutoDestroyAt string   `form:"destroyAt" json:"destroyAt" binding:"" `                          // 自动销毁时间， 0: 不自动销毁, 12h: 12小时,	1d：一天,3d: 三天	1w: 一周，7天	15d: 半个月，15天	1m: 一个月，28/29/30/31天根据不同月份	xxxx-xx-xx xx:xx：指定时间（选择指定时间时出现时间选择框），格式：年-月-日 时:分

	TaskType string `form:"taskType" json:"taskType" binding:"required" enums:"plan,apply,destroy"` // 环境创建后触发的任务步骤，plan计划,apply部署,destroy销毁资源
	Targets  string `form:"targets" json:"targets" binding:""`                                      // Terraform target 参数列表
	RunnerId string `form:"runnerId" json:"runnerId" binding:""`                                    // 环境默认部署通道
	Revision string `form:"revision" json:"revision" binding:""`                                    // 分支/标签
	Timeout  int    `form:"timeout" json:"timeout" binding:""`                                      // 部署超时时间（单位：秒）

	Variables []models.VariableBody `form:"variables" json:"variable" binding:""` // 自定义变量列表，该变量列表会覆盖现有的变量

	TfVarsFile   string `form:"tfVarsFile" json:"tfVarsFile" binding:""`     // Terraform tfvars 变量文件路径
	PlayVarsFile string `form:"playVarsFile" json:"playVarsFile" binding:""` // Ansible playbook 变量文件路径
	Playbook     string `form:"playbook" json:"playbook" binding:""`         // Ansible playbook 入口文件路径
}

type ArchiveEnvForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略

	Archived bool `form:"archived" json:"archived" binding:"required" enums:"true,false"` // 归档状态
}

type SearchEnvForm struct {
	PageForm

	Q        string `form:"q" json:"q" binding:""`                                                   // 环境名称，支持模糊查询
	Status   string `form:"status" json:"status" enums:"active,deploying,approving,failed,inactive"` // 环境状态
	Archived string `form:"archived" json:"archived" enums:"true,false,all"`                         // 归档状态，默认返回未归档环境
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
	PageForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}
