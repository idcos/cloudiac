package forms

import "cloudiac/portal/models"

type CreateAccessTokenForm struct {
	PageForm
	TplGuid string `json:"tplGuid" form:"tplGuid" `
	Action  string `json:"action" form:"action" `
}

type UpdateAccessTokenForm struct {
	PageForm
	Id     models.Id `form:"id" json:"id" binding:"required"`
	Action string    `json:"action" form:"action" `
}

type SearchAccessTokenForm struct {
	PageForm
	TplGuid string `json:"tplGuid" form:"tplGuid" `
}

type DeleteAccessTokenForm struct {
	PageForm
	Id models.Id `form:"id" json:"id" binding:"required"`
}

type DetailAccessTokenForm struct {
	PageForm
	Id models.Id `form:"id" json:"id" binding:"required"`
}
