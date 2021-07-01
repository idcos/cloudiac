package forms

import "cloudiac/portal/models"

type CreateAccessTokenForm struct {
	BaseForm
	TplGuid string `json:"tplGuid" form:"tplGuid" `
	Action  string `json:"action" form:"action" `
}

type UpdateAccessTokenForm struct {
	BaseForm
	Id     models.Id `form:"id" json:"id" binding:"required"`
	Action string    `json:"action" form:"action" `
}

type SearchAccessTokenForm struct {
	BaseForm
	TplGuid string `json:"tplGuid" form:"tplGuid" `
}

type DeleteAccessTokenForm struct {
	BaseForm
	Id models.Id `form:"id" json:"id" binding:"required"`
}

type DetailAccessTokenForm struct {
	BaseForm
	Id models.Id `form:"id" json:"id" binding:"required"`
}
