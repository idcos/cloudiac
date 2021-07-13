package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
)

type Organization struct {
	ctrl.BaseController
}

// Create 创建组织
// @Tags 组织
// @Summary 创建组织
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.CreateOrganizationForm true "parameter"
// @router /orgs [post]
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
// @Summary 组织查询
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param form query forms.SearchOrganizationForm true "parameter"
// @router /orgs [get]
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
// @Summary 组织信息编辑
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param orgId path string true "组织ID"
// @Param form formData forms.UpdateOrganizationForm true "parameter"
// @router /orgs/{orgId} [put]
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
// @Summary 删除组织（不支持）
// @Description 删除组织接口，目前不支持组织删除。
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param orgId path string true "组织ID"
// @Param form formData forms.DeleteOrganizationForm true "parameter"
// @router /orgs/{orgId} [delete]
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
// @Summary 组织信息详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param orgId path string true "组织ID"
// @router /orgs/{orgId} [get]
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
// @Summary 启用/禁用组织
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param orgId path string true "组织ID"
// @Param form formData forms.DisableOrganizationForm true "parameter"
// @router /orgs/{orgId}/status [put]
// @Success 200 {object} ctx.JSONResult{result=models.Organization}
func (Organization) ChangeOrgStatus(c *ctx.GinRequestCtx) {
	form := forms.DisableOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ChangeOrgStatus(c.ServiceCtx(), &form))
}

// AddUserToOrg 添加用户到组织
// @Tags 组织
// @Summary 添加用户到组织
// @Description 将一个或多个用户添加到组织里面。操作人需要拥有组织管理权限。
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param orgId path string true "组织ID"
// @Param form formData forms.AddUserOrgRelForm true "parameter"
// @router /orgs/{orgId}/users [put]
// @Success 200 {object} ctx.JSONResult{result=apps.UserWithRoleResp}
func (Organization) AddUserToOrg(c *ctx.GinRequestCtx) {
	form := forms.AddUserOrgRelForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.AddUserOrgRel(c.ServiceCtx(), &form))
}

// RemoveUserForOrg 从组织移除用户
// @Tags 组织
// @Summary 从组织移除用户
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param orgId path string true "组织ID"
// @Param userId path string true "用户ID"
// @Param form formData forms.DeleteUserOrgRelForm true "parameter"
// @router /orgs/{orgId}/users/{userId} [delete]
// @Success 200 {object} ctx.JSONResult{}
func (Organization) RemoveUserForOrg(c *ctx.GinRequestCtx) {
	form := forms.DeleteUserOrgRelForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteUserOrgRel(c.ServiceCtx(), &form))
}

// UpdateUserOrgRel 编辑用户组织角色
// @Tags 组织
// @Summary 编辑用户组织角色
// @Description 修改用户在组织中的角色。操作人需要拥有组织管理权限。
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param orgId path string true "组织ID"
// @Param userId path string true "用户ID"
// @Param form formData forms.UpdateUserOrgRelForm true "parameter"
// @router /orgs/{orgId}/users/{userId}/role [put]
// @Success 200 {object} ctx.JSONResult{result=apps.UserWithRoleResp}
func (Organization) UpdateUserOrgRel(c *ctx.GinRequestCtx) {
	form := forms.UpdateUserOrgRelForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateUserOrgRel(c.ServiceCtx(), &form))
}

/**
// SearchUser 查询组织用户列表
// @Tags 用户
// @Summary 用户查询
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param orgId path string true "组织ID"
// @Param Iac-Org-Id header string true "组织ID"
// @Param form query forms.SearchUserForm true "parameter"
// @router /orgs/{orgId}/users [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.User}}
*/
func (Organization) SearchUser(c *ctx.GinRequestCtx) {
	form := forms.SearchUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.ServiceCtx().OrgId = models.Id(c.Param("orgId"))
	c.JSONResult(apps.SearchUser(c.ServiceCtx(), &form))
}
