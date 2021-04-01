package handlers

import (
	"cloudiac/apps"
	"cloudiac/consts/e"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type Template struct {
	ctrl.BaseController
}

func (Template) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateTemplateForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateTemplate(c.ServiceCtx(), form))
}

func (Template) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchTemplate(c.ServiceCtx(), &form))
}

func (Template) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateTemplate(c.ServiceCtx(), &form))
}

func (Template) Delete(c *ctx.GinRequestCtx) {
	//form := forms.DeleteUserForm{}
	//if err := c.Bind(&form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.DeleteUser(c.ServiceCtx(), &form))
	c.JSONError(e.New(e.NotImplement))
}

func (Template) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DetailTemplate(c.ServiceCtx(), &form))
}

func (Template) Overview(c *ctx.GinRequestCtx) {
	form := forms.OverviewTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.OverviewTemplate(c.ServiceCtx(), &form))
}


