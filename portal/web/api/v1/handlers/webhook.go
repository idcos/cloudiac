package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type AccessToken struct {
	ctrl.BaseController
}

func (AccessToken) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateAccessTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateAccessToken(c.ServiceCtx(), form))
}

func (AccessToken) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchAccessTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchAccessToken(c.ServiceCtx(), &form))
}

func (AccessToken) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateAccessTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateAccessToken(c.ServiceCtx(), &form))
}

func (AccessToken) Delete(c *ctx.GinRequestCtx) {
	form := forms.DeleteAccessTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteAccessToken(c.ServiceCtx(), &form))
}

func (AccessToken) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailAccessTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DetailAccessToken(c.ServiceCtx(), &form))
}
