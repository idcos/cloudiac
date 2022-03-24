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
	PageForm
	Name         string   `form:"name" json:"name" binding:"required,gte=2,lte=32"`
	Description  string   `form:"description" json:"description" binding:"max=255"`
	Params       []Params `form:"params" json:"params" binding:"required"`
	CtServiceIds []string `form:"ctServiceIds" json:"ctServiceIds"`
}

type UpdateResourceAccountForm struct {
	PageForm
	Id           models.Id `form:"id" json:"id" binding:"required"`
	Name         string    `form:"name" json:"name" binding:"omitempty,gte=2,lte=32"`
	Description  string    `form:"description" json:"description" binding:"max=255"`
	Params       []Params  `form:"params" json:"params"`
	Status       string    `form:"status" json:"status" binding:"omitempty,oneof=enable,disable"`
	CtServiceIds []string  `form:"ctServiceIds" json:"ctServiceIds"`
}

type SearchResourceAccountForm struct {
	PageForm

	Q      string `form:"q" json:"q" binding:""`
	Status string `form:"status" json:"status" binding:"omitempty,oneof=enable disable"`
}

type DeleteResourceAccountForm struct {
	PageForm
	Id models.Id `uri:"id" json:"id" binding:"required"`
}
