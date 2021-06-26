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

// Create 创建组织
// @Summary 创建组织
// @Description 创建组织
// @Tags 组织
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.CreateOrganizationForm true "云模板信息"
// @Success 200 {object} models.Organization
// @Router /org/create [post]
func (Organization) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateOrganizationForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateOrganization(c.ServiceCtx(), form))
}

// Search 查询组织列表
// @Summary 查询组织列表
// @Description 查询组织列表
// @Tags 组织
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param q query string false "模糊搜索"
// @Param status query string false "组织状态"
// @Success 200 {object} apps.searchOrganizationResp
// @Router /org/search [get]
func (Organization) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchOrganization(c.ServiceCtx(), &form))
}

// Update 修改组织信息
// @Summary 修改组织信息
// @Description 修改组织信息
// @Tags 组织
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.UpdateOrganizationForm true "组织信息"
// @Success 200 {object} models.Organization
// @Router /org/update [put]
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

// Detail 查询组织详情
// @Summary 查询组织详情
// @Description 查询组织详情
// @Tags 组织
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param id query int true "组织id"
// @Success 200 {object} apps.organizationDetailResp
// @Router /org/detail [get]
func (Organization) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.OrganizationDetail(c.ServiceCtx(), &form))
}

// ChangeOrgStatus 组织状态修改
// @Summary 组织状态修改
// @Description 组织状态修改
// @Tags 组织
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.DisableOrganizationForm true "组织状态信息"
// @Success 200 {object} models.Organization
// @Router /org/status/update [put]
func (Organization) ChangeOrgStatus(c *ctx.GinRequestCtx) {
	form := forms.DisableOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ChangeOrgStatus(c.ServiceCtx(), &form))
}