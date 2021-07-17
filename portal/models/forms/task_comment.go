package forms

import "cloudiac/portal/models"

type CreateTaskCommentForm struct {
	PageForm

	Id      models.Id `url:"id" json:"id" form:"id" binding:""`
	Comment string    `json:"comment" form:"comment" binding:"required"`
}

type SearchTaskCommentForm struct {
	PageForm
	Id models.Id `url:"id" json:"id" form:"id" `
}
