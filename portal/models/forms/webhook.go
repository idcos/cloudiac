package forms

import "cloudiac/portal/models"

type CreateApiTriggerForm struct {
	PageForm
	TplGuid string `json:"tplGuid" form:"tplGuid" `
	Action  string `json:"action" form:"action" `
}

type UpdateApiTriggerForm struct {
	PageForm
	Id     models.Id `form:"id" json:"id" binding:"required"`
	Action string    `json:"action" form:"action" `
}

type SearchApiTriggerForm struct {
	PageForm
	TplGuid string `json:"tplGuid" form:"tplGuid" `
}

type DeleteApiTriggerForm struct {
	PageForm
	Id models.Id `form:"id" json:"id" binding:"required"`
}

type DetailApiTriggerForm struct {
	PageForm
	Id models.Id `form:"id" json:"id" binding:"required"`
}
