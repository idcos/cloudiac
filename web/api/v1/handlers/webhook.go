package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type Webhook struct {
	ctrl.BaseController
}

func (Webhook) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateWebhookForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateWebhook(c.ServiceCtx(), form))
}

func (Webhook) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchWebhookForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchWebhook(c.ServiceCtx(), &form))
}

func (Webhook) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateWebhookForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateWebhook(c.ServiceCtx(), &form))
}

func (Webhook) Delete(c *ctx.GinRequestCtx) {
	form := forms.DeleteWebhookForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteWebhook(c.ServiceCtx(), &form))
}

func (Webhook) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailWebhookForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DetailWebhook(c.ServiceCtx(), &form))
}
