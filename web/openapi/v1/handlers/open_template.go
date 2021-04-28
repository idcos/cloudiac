package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctx"
)

func OpenTemplateSearch(c *ctx.GinRequestCtx) {
	c.JSONOpenResultList(apps.OpenSearchTemplate(c.ServiceCtx()))
}
