package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type ResourceAccount struct {
	ctrl.BaseController
}

func (ResourceAccount) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchResourceAccount(c.ServiceCtx(), form))
}

func (ResourceAccount) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateResourceAccount(c.ServiceCtx(), form))
}

func (ResourceAccount) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteResourceAccount(c.ServiceCtx(), form))
}

func (ResourceAccount) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateResourceAccount(c.ServiceCtx(), form))
}
