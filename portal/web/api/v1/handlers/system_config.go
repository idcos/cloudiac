package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type SystemConfig struct {
	ctrl.BaseController
}

func (SystemConfig) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateOrganizationForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateOrganization(c.ServiceCtx(), form))
}

func (SystemConfig) Search(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.SearchSystemConfig(c.ServiceCtx()))
}

func (SystemConfig) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateSystemConfigForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateSystemConfig(c.ServiceCtx(), &form))
}
