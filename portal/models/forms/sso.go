// Copyright 2021 CloudJ Company Limited. All rights reserved.

package forms

type VerifySsoTokenForm struct {
	BaseForm

	Token string `json:"token" form:"token" binding:"required"`
}
