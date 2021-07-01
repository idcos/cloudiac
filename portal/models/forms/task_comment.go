package forms

import "cloudiac/portal/models"

type CreateTaskCommentForm struct {
	BaseForm

	TaskId  models.Id `json:"taskId" form:"taskId" binding:"required"`
	Comment string    `json:"comment" form:"comment" binding:"required"`
}

type SearchTaskCommentForm struct {
	BaseForm
	TaskId models.Id `json:"taskId" form:"taskId" `
}
