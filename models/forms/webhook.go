package forms

type CreateAccessTokenForm struct {
	BaseForm
	TplGuid string `json:"tplGuid" form:"tplGuid" `
	Action  string `json:"action" form:"action" `
}

type UpdateAccessTokenForm struct {
	BaseForm
	Id     uint   `form:"id" json:"id" binding:"required"`
	Action string `json:"action" form:"action" `
}

type SearchAccessTokenForm struct {
	BaseForm
	TplGuid string `json:"tplGuid" form:"tplGuid" `
}

type DeleteAccessTokenForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}

type DetailAccessTokenForm struct {
	BaseForm
	Id uint `form:"id" json:"id" binding:"required"`
}
