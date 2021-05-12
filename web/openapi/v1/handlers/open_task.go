package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

func TaskCreate(c *ctx.GinRequestCtx) {
	form := forms.CreateTaskOpenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONOpenResultItem(apps.CreateTaskOpen(c.ServiceCtx(), form))
}
