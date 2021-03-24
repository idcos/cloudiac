package handlers

import (
	"cloudiac/apps"
	"cloudiac/consts/e"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type Organization struct {
	ctrl.BaseController
}

func (Organization) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateOrganizationForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateOrganization(c.ServiceCtx(), form))
}

func (Organization) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchOrganization(c.ServiceCtx(), &form))
}

func (Organization) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateOrganization(c.ServiceCtx(), &form))
}

func (Organization) Delete(c *ctx.GinRequestCtx) {
	//form := forms.DeleteUserForm{}
	//if err := c.Bind(&form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.DeleteUser(c.ServiceCtx(), &form))
	c.JSONError(e.New(e.NotImplement))
}

func (Organization) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.OrganizationDetail(c.ServiceCtx(), &form))
}

func (Organization) ChangeOrgStatus(c *ctx.GinRequestCtx) {
	form := forms.DisableOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ChangeOrgStatus(c.ServiceCtx(), &form))
}
