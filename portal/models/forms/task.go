// Copyright 2021 CloudJ Company Limited. All rights reserved.

package forms

import "cloudiac/portal/models"

type CreateTaskForm struct {
	BaseForm

	Name          string    `json:"name" form:"name" `
	CtServiceIp   string    `json:"ctServiceIp" form:"ctServiceIp" binding:"required"`
	CtServicePort uint      `json:"ctServicePort" form:"ctServicePort" binding:"required"`
	CtServiceId   string    `json:"ctServiceId" form:"ctServiceId" binding:"required"`
	TemplateId    models.Id `json:"templateId" form:"templateId" binding:"required"`
	TaskType      string    `json:"taskType" form:"taskType" binding:"required"`
}

type DetailTaskForm struct {
	BaseForm
	Id models.Id `uri:"id" form:"id" json:"id" swaggerignore:"true"` // 任务ID，swagger 参数通过 param path 指定，这里忽略
}

type DetailTaskStepForm struct {
	PageForm
	TaskId models.Id `uri:"id" form:"id" json:"id" swaggerignore:"true"` // 任务ID，swagger 参数通过 param path 指定，这里忽略
}

type TaskLogForm struct {
	BaseForm
	Id       models.Id `uri:"id" form:"id" json:"id" swaggerignore:"true"`             // 任务ID，swagger 参数通过 param path 指定，这里忽略
	StepType string    `form:"stepType" json:"stepType"`                               // 步骤名称
	StepId   models.Id `uri:"stepId" form:"stepId" json:"stepId" swaggerignore:"true"` // 任务步骤步骤ID
}

type SearchTaskForm struct {
	NoPageSizeForm

	EnvId models.Id `json:"envId" form:"envId" binding:"required"` // 环境ID
}

type LastTaskForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type TaskStateListForm struct {
	PageForm
	TemplateId models.Id `json:"templateId" form:"templateId" `
	TaskId     models.Id `json:"taskId" form:"taskId" binding:"required"`
}

type UpdateTaskForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 任务ID，swagger 参数通过 param path 指定，这里忽略

	Name        string `form:"name" json:"name" binding:""`                      // 任务名称
	Description string `form:"description" json:"description" binding:"max=255"` // 任务描述
	RunnerId    string `form:"runnerId" json:"runnerId" binding:""`              // 任务默认部署通道
	Status      string `form:"status" json:"status" enums:"enable,disable"`      // 任务状态
}

const (
	TaskActionApproved = "approved"
	TaskActionRejected = "rejected"
)

type ApproveTaskForm struct {
	BaseForm

	Id     models.Id `uri:"id" json:"id" swaggerignore:"true"`                                  // 任务ID，swagger 参数通过 param path 指定，这里忽略
	Action string    `form:"action" json:"action" binding:"required" enums:"approved,rejected"` // 审批动作：approved通过, rejected驳回
}

type SearchEnvTasksForm struct {
	NoPageSizeForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type SearchTaskResourceForm struct {
	NoPageSizeForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 任务ID，swagger 参数通过 param path 指定，这里忽略
	Q  string    `form:"q" json:"q" binding:""`            // 资源名称，支持模糊查询
}

type ResourceDetailForm struct {
	BaseForm

	Id         models.Id `uri:"id" json:"id" swaggerignore:"true"`                 // 环境ID，swagger 参数通过 param path 指定，这里忽略
	ResourceId models.Id `uri:"resourceId" json:"resourceId" swaggerignore:"true"` // 部署成功后后资源ID
}

type GetTaskStepLogForm struct {
	BaseForm
	Id     models.Id `uri:"id" json:"id"`         // 任务Id
	StepId models.Id `uri:"stepId" json:"stepId"` //步骤ID
}
