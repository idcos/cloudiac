package forms

import "cloudiac/portal/models"

type SearchSystemConfigForm struct {
	PageForm

	Q string `form:"q" json:"q" binding:""`
}

type UpdateSystemConfigForm struct {
	PageForm
	Id          models.Id `form:"id" json:"id" binding:""`
	Name        string    `form:"name" json:"name" binding:""`
	Value       string    `form:"value" json:"value" binding:"required"`
	Description string    `form:"description" json:"description"`
}
