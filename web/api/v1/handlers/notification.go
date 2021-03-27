package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

func (Organization) ListNotificationCfgs(c *ctx.GinRequestCtx) {
	orgId, _ := c.QueryInt("orgId")
	c.JSONResult(apps.ListNotificationCfgs(c.ServiceCtx(), orgId))
}

func (Organization) CreateNotificationCfgs(c *ctx.GinRequestCtx) {
	form := &forms.CreateNotificationCfgForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateNotificationCfg(c.ServiceCtx(), form))
}

func (Organization) DeleteNotificationCfgs(c *ctx.GinRequestCtx) {
	cfgId, _ := c.QueryInt("notificationId")
	c.JSONResult(apps.DeleteNotificationCfg(c.ServiceCtx(), cfgId))
}

func (Organization) UpdateNotificationCfgs(c *ctx.GinRequestCtx) {
	form := &forms.UpdateNotificationCfgForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateNotificationCfg(c.ServiceCtx(), form))
}
