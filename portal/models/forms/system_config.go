// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

type SearchSystemConfigForm struct {
	PageForm

	Q string `form:"q" json:"q" binding:""`
}

type UpdateSystemConfigForm struct {
	BaseForm
	SystemCfg []SystemCfg `json:"systemCfg" form:"systemCfg" `
}

type SystemCfg struct {
	Name        string `form:"name" json:"name" binding:"required"`
	Value       string `form:"value" json:"value" binding:"required"`
	Description string `form:"description" json:"description"`
}

type RegistryAddrForm struct {
	BaseForm
	RegistryAddr string `form:"registryAddr" json:"registryAddr"`
}
