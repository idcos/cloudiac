package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Token struct {
	ctrl.BaseController
}

func (Token) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateToken(c.ServiceCtx(), form))
}

func (Token) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchToken(c.ServiceCtx(), form))
}

func (Token) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateToken(c.ServiceCtx(), form))
}

func (Token) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteToken(c.ServiceCtx(), form))
}
