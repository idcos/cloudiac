package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type Notification struct {
	ctrl.BaseController
}

func (Notification) Search(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.ListNotificationCfgs(c.ServiceCtx()))
}

func (Notification) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateNotificationCfgForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateNotificationCfg(c.ServiceCtx(), form))
}

func (Notification) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteNotificationCfgForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteNotificationCfg(c.ServiceCtx(), form.UserId))
}

func (Notification) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateNotificationCfgForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateNotificationCfg(c.ServiceCtx(), form))
}
