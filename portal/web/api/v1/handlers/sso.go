package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// TODO: swagger docs
// GenerateSsoToken 生成 SSO token
func GenerateSsoToken(c *ctx.GinRequest) {
	c.JSONResult(apps.GenerateSsoToken(c.Service()))
}

// TODO: swagger docs
// VerifySsoToken 验证 SSO token
func VerifySsoToken(c *ctx.GinRequest) {
	form := forms.VerifySsoTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}

	c.JSONResult(apps.VerifySsoToken(c.Service(), &form))
}
