// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type CreateTagForm struct {
	BaseForm

	Key        string    `json:"key" form:"key" `
	Value      string    `json:"value" form:"value" `
	ObjectType string    `json:"objectType" form:"objectType" `
	ObjectId   models.Id `json:"objectId" form:"objectId" `
}

type DeleteTagsForm struct {
	BaseForm

	ObjectType string    `json:"objectType" form:"objectType" `
	ObjectId   models.Id `json:"objectId" form:"objectId" `
	KeyId      models.Id `json:"keyId" form:"keyId" `
	ValueId    models.Id `json:"valueId" form:"valueId" `
}

type UpdateTagsForm struct {
	BaseForm

	KeyId      models.Id `json:"keyId" form:"keyId" `
	ValueId    models.Id `json:"valueId" form:"valueId" `
	Key        string    `json:"key" form:"key" `
	Value      string    `json:"value" form:"value" `
	ObjectType string    `json:"objectType" form:"objectType" `
	ObjectId   models.Id `json:"objectId" form:"objectId" `
}

type SearchTagsForm struct {
	PageForm

	Q          string    `json:"q" form:"q" `
	ObjectType string    `json:"objectType" form:"objectType" binding:"required"`
	ObjectId   models.Id `json:"objectId" form:"objectId" binding:"required"`
}

type SearchEnvTagsForm struct {
	PageForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=env-,max=32"` // 环境ID，swagger 参数通过 param path 指定，这里忽略
	Q  string    `json:"q" form:"q" `
}
