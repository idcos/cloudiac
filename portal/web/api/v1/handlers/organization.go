package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Organization struct {
	ctrl.BaseController
}

func (Organization) Create(c *ctx.GinRequestCtx) {
	form := forms.CreateOrganizationForm{}
	form.Bind(nil)
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.CreateOrganization(c.ServiceCtx(), &form))
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
	param := forms.UpdateOrganizationParam{}
	if err := c.BindUri(&param); err != nil {
		// 如果 uri 参数不对不应该进到这里
		c.Logger().Panic(err)
		return
	}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateOrganization(c.ServiceCtx(), param.Id, &form))
}

func (Organization) Delete(c *ctx.GinRequestCtx) {
	// 组织不允许删除
	c.JSONError(e.New(e.NotImplement))
}

func (Organization) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.OrganizationDetail(c.ServiceCtx(), form))
}

func (Organization) ChangeOrgStatus(c *ctx.GinRequestCtx) {
	form := forms.DisableOrganizationForm{}
	param := forms.UpdateOrganizationParam{}
	if err := c.BindUri(&param); err != nil {
		// 如果 uri 参数不对不应该进到这里
		c.Logger().Panic(err)
		return
	}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ChangeOrgStatus(c.ServiceCtx(), param.Id, &form))
}
