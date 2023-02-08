// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package forms

type ClearProviderCacheForm struct {
	BaseForm

	Source  string `json:"source" form:"source" binding:"required"`
	Version string `json:"version" form:"version" binding:"required"`
}
