// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

type SearchSystemConfigForm struct {
	PageForm

	Q string `form:"q" json:"q" binding:""`
}

type UpdateSystemConfigForm struct {
	BaseForm
	SystemCfg []SystemCfg `json:"systemCfg" form:"systemCfg" binding:"required,dive,required" `
}

type SystemCfg struct {
	Name        string `form:"name" json:"name" binding:"required,gte=2,lte=255"`
	Value       string `form:"value" json:"value" binding:"required,gte=2,lte=32"`
	Description string `form:"description" json:"description" binding:"max=255"`
}

type RegistryAddrForm struct {
	BaseForm
	RegistryAddr string `form:"registryAddr" json:"registryAddr" binding:""`
}
