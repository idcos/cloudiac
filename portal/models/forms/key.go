// Copyright 2021 CloudJ Company Limited. All rights reserved.

package forms

import (
	"cloudiac/portal/models"
)

type CreateKeyForm struct {
	BaseForm

	Name string `json:"name" form:"name"` // 密钥名称
	Key  string `json:"key" form:"key"`   // 密钥内容
}

type SearchKeyForm struct {
	PageForm

	Q string `form:"q" json:"q" binding:""` // 密钥名称，支持模糊搜索
}

type DetailKeyForm struct {
	BaseForm

	Id models.Id `uri:"id" form:"id" json:"id" binding:"" swaggerignore:"true"` // 密钥ID
}

type UpdateKeyForm struct {
	BaseForm

	Id   models.Id `uri:"id" form:"id" json:"id" binding:"" swaggerignore:"true"` // 密钥ID
	Name string    `json:"name" form:"name"`                                      // 名称
}

type DeleteKeyForm struct {
	BaseForm

	Id models.Id `uri:"id" form:"id" json:"id" binding:"" swaggerignore:"true"` // 密钥ID
}
