package forms

import "cloudiac/portal/models"

type CreateTokenForm struct {
	BaseForm

	Description string `form:"description" json:"description" binding:""`
}

type UpdateTokenForm struct {
	BaseForm
	Id          models.Id `form:"id" json:"id" binding:"required"`
	Status      string    `form:"status" json:"status" binding:"required"`
	Description string    `form:"description" json:"description" binding:""`
}

type SearchTokenForm struct {
	BaseForm
	Q      string `form:"q" json:"q" binding:""`
	Status string `form:"status" json:"status"`
}

type DeleteTokenForm struct {
	BaseForm
	Id models.Id `form:"id" json:"id" binding:"required"`
}
