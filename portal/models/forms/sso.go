// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

type VerifySsoTokenForm struct {
	BaseForm

	Token string `json:"token" form:"token" binding:"required"`
}
