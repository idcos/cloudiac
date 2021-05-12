package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

func OpenTemplateSearch(c *ctx.GinRequestCtx) {
	c.JSONOpenResultList(apps.OpenSearchTemplate(c.ServiceCtx()))
}

func TemplateDetail(c *ctx.GinRequestCtx) {
	form := forms.OpenApiDetailTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONOpenResultItem(apps.OpenDetailTemplate(c.ServiceCtx(), form.Guid))
}
