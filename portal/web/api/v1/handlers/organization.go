package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Organization struct {
	ctrl.BaseController
}

// Create 创建组织
// @Tags 组织
// @Description 创建组织接口
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Param Authorization header string true "认证令牌"
// @Param form formData forms.CreateOrganizationForm true "parameter"
// @router /api/v1/orgs [post]
// @Success 200 {object} ctx.JSONResult{result=models.Organization}
func (Organization) Create(c *ctx.GinRequestCtx) {
	form := forms.CreateOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.CreateOrganization(c.ServiceCtx(), &form))
}

// Search 组织查询
// @Tags 组织
// @Description 组织查询接口
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string true "认证令牌"
// @Param form query forms.SearchOrganizationForm true "parameter"
// @router /api/v1/orgs [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Organization}}
func (Organization) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchOrganization(c.ServiceCtx(), &form))
}

// Update 组织编辑
// @Tags 组织
// @Description 组织信息编辑接口
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Param Authorization header string true "认证令牌"
// @Param orgId path string true "组织ID"
// @Param form formData forms.UpdateOrganizationForm true "parameter"
// @router /api/v1/orgs/{orgId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.Organization}
func (Organization) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateOrganization(c.ServiceCtx(), &form))
}

// Delete 删除组织（不支持）
// @Tags 组织
// @Description 删除组织接口，目前不支持组织删除。
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Param Authorization header string true "认证令牌"
// @Param orgId path string true "组织ID"
// @Param form formData forms.DeleteOrganizationForm true "parameter"
// @router /api/v1/orgs/{orgId} [delete]
// @Success 501 {object} ctx.JSONResult
func (Organization) Delete(c *ctx.GinRequestCtx) {
	form := forms.DeleteOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteOrganization(c.ServiceCtx(), &form))
}

// Detail 组织信息详情
// @Tags 组织
// @Description 组织信息详情查询接口
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Authorization header string true "认证令牌"
// @Param orgId path string true "组织ID"
// @router /api/v1/orgs/{orgId} [get]
// @Success 200 {object} ctx.JSONResult{result=apps.organizationDetailResp}
func (Organization) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.OrganizationDetail(c.ServiceCtx(), form))
}

// ChangeOrgStatus 启用/禁用组织
// @Tags 组织
// @Description 组织状态启用、禁用状态修改接口
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Param Authorization header string true "认证令牌"
// @Param orgId path string true "组织ID"
// @Param form formData forms.DisableOrganizationForm true "parameter"
// @router /api/v1/orgs/{orgId}/status [put]
// @Success 200 {object} ctx.JSONResult{result=models.Organization}
func (Organization) ChangeOrgStatus(c *ctx.GinRequestCtx) {
	form := forms.DisableOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ChangeOrgStatus(c.ServiceCtx(), &form))
}
