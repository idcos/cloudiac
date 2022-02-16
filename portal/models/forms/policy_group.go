// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

type SearchRegistryPgForm struct {
	PageForm

	Q string `json:"q" form:"q"`
}

type SearchRegistryPgVersForm struct {
	BaseForm

	Namespace string `json:"ns" form:"ns" binding:"required"` // policy namespace
	GroupName string `json:"gn" form:"gn" binding:"required"` // policy groupname
}
