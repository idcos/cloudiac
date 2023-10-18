// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type CreateTaskForm struct {
	BaseForm

	Name          string    `json:"name" form:"name" binding:"max=255"`
	CtServiceIp   string    `json:"ctServiceIp" form:"ctServiceIp" binding:"required"`
	CtServicePort uint      `json:"ctServicePort" form:"ctServicePort" binding:"required"`
	CtServiceId   string    `json:"ctServiceId" form:"ctServiceId" binding:"required"`
	TemplateId    models.Id `json:"templateId" form:"templateId" binding:"required"`
	TaskType      string    `json:"taskType" form:"taskType" binding:"required"`
}

type DetailTaskForm struct {
	BaseForm
	Id models.Id `uri:"id" form:"id" json:"id" binding:"required,startswith=run-,max=32" swaggerignore:"true"` // 任务ID，swagger 参数通过 param path 指定，这里忽略
}

type DetailTaskStepForm struct {
	PageForm
	TaskId models.Id `uri:"id" form:"id" json:"id" binding:"required,startswith=run-,max=32" swaggerignore:"true"` // 任务ID，swagger 参数通过 param path 指定，这里忽略
}

type TaskLogForm struct {
	BaseForm
	Id       models.Id `uri:"id" form:"id" json:"id" binding:"required,startswith=run-,max=32" swaggerignore:"true"` // 任务ID，swagger 参数通过 param path 指定，这里忽略
	StepType string    `form:"stepType" json:"stepType" binding:""`                                                  // 步骤名称
	StepId   models.Id `uri:"stepId" form:"stepId" json:"stepId" swaggerignore:"true"`                               // 任务步骤步骤ID
}

type SearchTaskForm struct {
	NoPageSizeForm

	EnvId    models.Id `json:"envId" form:"envId" binding:"required,startswith=run-,max=32"` // 环境ID
	TaskType string    `form:"taskType" json:"taskType" binding:""`                          // 任务类型
	Source   string    `form:"source" json:"source"`                                         // 触发类型
	User     string    `form:"user" json:"user"`                                             // 可根据执行人姓名或邮箱模糊查询
}

type LastTaskForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=run-,max=32"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type TaskStateListForm struct {
	PageForm
	TemplateId models.Id `json:"templateId" form:"templateId" `
	TaskId     models.Id `json:"taskId" form:"taskId" binding:"required,startswith=run-,max=32"`
}

type UpdateTaskForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=run-,max=32"` // 任务ID，swagger 参数通过 param path 指定，这里忽略

	Name        string `form:"name" json:"name" binding:"omitempty,gte=2,lte=64"`                                    // 任务名称
	Description string `form:"description" json:"description" binding:"max=255"`                                     // 任务描述
	RunnerId    string `form:"runnerId" json:"runnerId" binding:"max=255"`                                           // 任务默认部署通道
	Status      string `form:"status" json:"status" binding:"omitempty,oneof=enable disable" enums:"enable,disable"` // 任务状态
}

const (
	TaskActionApproved = "approved"
	TaskActionRejected = "rejected"
)

type ApproveTaskForm struct {
	BaseForm

	Id     models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=run-,max=32"`                // 任务ID，swagger 参数通过 param path 指定，这里忽略
	Action string    `form:"action" json:"action" binding:"required,oneof=approved rejected" enums:"approved,rejected"` // 审批动作：approved通过, rejected驳回
}

type AbortTaskForm struct {
	BaseForm

	TaskId models.Id `uri:"id" json:"taskId" swaggerignore:"true"` // 任务ID，swagger 参数通过 param path 指定，这里忽略
}

type SearchEnvTasksForm struct {
	NoPageSizeForm

	Id       models.Id `uri:"id" json:"id" swaggerignore:"true" bingding:"omitempty,startswith=env-,max=32"`                                        // 环境ID，swagger 参数通过 param path 指定，这里忽略
	TaskType string    `form:"taskType" json:"taskType" binding:"omitempty,oneof=plan apply destroy scan"`                                          // 任务类型
	Source   string    `form:"source" json:"source" binding:"omitempty,oneof=manual driftPlan driftApply webhookPlan webhookApply autoDestroy api"` // 触发类型
	User     string    `form:"user" json:"user"`                                                                                                    // 可根据执行人姓名或邮箱模糊查询
}

type SearchTaskResourceForm struct {
	NoPageSizeForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=run-,max=32"` // 任务ID，swagger 参数通过 param path 指定，这里忽略
	Q  string    `form:"q" json:"q" binding:""`                                                      // 资源名称，支持模糊查询
}

type ResourceDetailForm struct {
	BaseForm

	Id         models.Id `uri:"id" json:"id" binding:"required,startswith=env-,max=32" swaggerignore:"true"`               // 环境ID，swagger 参数通过 param path 指定，这里忽略
	ResourceId models.Id `uri:"resourceId" json:"resourceId" binding:"required,startswith=r-,max=32" swaggerignore:"true"` // 部署成功后后资源ID
}

type GetTaskStepLogForm struct {
	BaseForm
	Id      models.Id `uri:"id" json:"id" binding:"required,startswith=run-,max=32"`          // 任务Id
	StepId  models.Id `uri:"stepId" json:"stepId" binding:"required,startswith=step-,max=32"` //步骤ID
	Number  int       `json:"number" form:"number"`                                           // 返回行数
	ShowAll bool      `json:"showAll" form:"showAll"`                                         // 是否展示所有
}

type ErrorStepLogForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" binding:"required,startswith=run-,max=32"` // 任务Id
}

type SearchTaskResourceGraphForm struct {
	BaseForm

	Id        models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=run-,max=32"` // 任务ID，swagger 参数通过 param path 指定，这里忽略
	Dimension string    `json:"dimension" form:"dimension" binding:"required"`                              // 资源名称，支持模糊查询
}
