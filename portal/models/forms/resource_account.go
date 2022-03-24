// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type Params struct {
	Id       string `json:"id" form:"id" `
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret *bool  `json:"isSecret"`
}

type CreateResourceAccountForm struct {
	BaseForm
	Name         string   `form:"name" json:"name" binding:"required,gte=2,lte=32"` // 资源账号名字
	Description  string   `form:"description" json:"description"`
	Params       []Params `form:"params" json:"params"`
	CtServiceIds []string `form:"ctServiceIds" json:"ctServiceIds"`
}

type UpdateResourceAccountForm struct {
	BaseForm
	Id           models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required"`
	Name         string    `form:"name" json:"name" binding:""`
	Description  string    `form:"description" json:"description"`
	Params       []Params  `form:"params" json:"params"`
	Status       string    `form:"status" json:"status"`
	CtServiceIds []string  `form:"ctServiceIds" json:"ctServiceIds"`
}

type SearchResourceAccountForm struct {
	PageForm

	Q      string `form:"q" json:"q" binding:""`                       // 资源账号名或者描述 支持模糊查询
	Status string `form:"status" json:"status" enums:"enable,disable"` // 资源账号状态
}

type DeleteResourceAccountForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" binding:"required"`
}
