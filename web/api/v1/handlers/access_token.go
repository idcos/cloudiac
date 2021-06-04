package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

func AccessTokenHandler(c *ctx.GinRequestCtx) {
	form := forms.AccessTokenHandler{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.AccessTokenHandler(c.ServiceCtx(),form))
}
