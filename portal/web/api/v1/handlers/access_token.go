package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

func AccessTokenHandler(c *ctx.GinRequestCtx) {
	form := forms.AccessTokenHandler{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.AccessTokenHandler(c.ServiceCtx(), form))
}
