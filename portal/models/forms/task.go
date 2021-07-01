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
	TaskId models.Id `json:"taskId" form:"taskId" binding:"required"`
}

type SearchTaskForm struct {
	BaseForm
	TemplateId models.Id `json:"templateId" form:"templateId" binding:"required"`
	Q          string    `form:"q" json:"q" binding:""`
	Status     string    `form:"status" json:"status"`
}

type LastTaskForm struct {
	BaseForm
	TemplateId models.Id `json:"templateId" form:"templateId" binding:"required"`
}

type TaskStateListForm struct {
	BaseForm
	TemplateId models.Id `json:"templateId" form:"templateId" `
	TaskId     models.Id `json:"taskId" form:"taskId" binding:"required"`
}
