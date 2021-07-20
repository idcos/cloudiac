package forms

import "cloudiac/portal/models"

type SearchSystemConfigForm struct {
	PageForm

	Q string `form:"q" json:"q" binding:""`
}

type UpdateSystemConfigForm struct {
	BaseForm
	Id          models.Id `uri:"id" form:"id" json:"id" binding:""`
	Name        string    `form:"name" json:"name" binding:""`
	Value       string    `form:"value" json:"value" binding:"required"`
	Description string    `form:"description" json:"description"`
}
