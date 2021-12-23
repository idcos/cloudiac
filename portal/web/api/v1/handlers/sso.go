package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
)

// TODO: swagger docs
func GenerateSsoToken(c *ctx.GinRequest) {
	c.JSONResult(apps.GenerateSsoToken(c.Service()))
}

// TODO: swagger docs
func VerifySsoToken(c *ctx.GinRequest) {
	form := apps.VerifySsoTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}

	c.JSONResult(apps.VerifySsoToken(c.Service(), &form))
}
