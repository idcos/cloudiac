package forms

type CreateWebhookForm struct {
	BaseForm
	TplGuid string `json:"tplGuid" form:"tplGuid" `
	TplId   uint   `json:"tplId" form:"tplId" `
	Action  string `json:"action" form:"action" `
}

type UpdateWebhookForm struct {
	BaseForm
	Id     uint   `form:"id" json:"id" binding:"required"`
	Action string `json:"action" form:"action" `
}

type SearchWebhookForm struct {
	BaseForm
	TplGuid string `json:"tplGuid" form:"tplGuid" `
	TplId   uint   `json:"tplId" form:"tplId" `
}

type DeleteWebhookForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}

type DetailWebhookForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}
