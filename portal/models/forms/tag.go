package forms

import "cloudiac/portal/models"

type UpdateTagForm struct {
	BaseForm

	Id    models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
	TagId models.Id `uri:"tagId" json:"tagId"`
	Value string    `json:"value" form:"value" `
}

type DeleteTagForm struct {
	BaseForm

	Id    models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
	TagId models.Id `uri:"tagId" json:"tagId"`
}

type CreateTagForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
}

type SearchTagForm struct {
	PageForm

	Q string `json:"q" form:"q" `
}
