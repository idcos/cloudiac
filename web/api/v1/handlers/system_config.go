package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
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
	form := forms.SearchSystemConfigForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchSystemConfig(c.ServiceCtx(), &form))
}

func (SystemConfig) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateSystemConfigForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateSystemConfig(c.ServiceCtx(), &form))
}
