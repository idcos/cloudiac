// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
)

type Organization struct {
	ctrl.GinController
}

// Create 创建组织
// @Tags 组织
// @Summary 创建组织
// @Description 需要管理员权限
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.CreateOrganizationForm true "parameter"
// @router /orgs [post]
// @Success 200 {object} ctx.JSONResult{result=models.Organization}
func (Organization) Create(c *ctx.GinRequest) {
	form := forms.CreateOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.CreateOrganization(c.Service(), &form))
}

// Search 组织列表查询
// @Tags 组织
// @Summary 组织列表查询
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param form query forms.SearchOrganizationForm true "parameter"
// @router /orgs [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.OrgDetailResp}}
func (Organization) Search(c *ctx.GinRequest) {
	form := forms.SearchOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchOrganization(c.Service(), &form))
}

// Update 组织编辑
// @Tags 组织
// @Summary 组织信息编辑
// @Description 需要管理员权限
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param orgId path string true "组织ID"
// @Param form formData forms.UpdateOrganizationForm true "parameter"
// @router /orgs/{orgId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.Organization}
func (Organization) Update(c *ctx.GinRequest) {
	form := forms.UpdateOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateOrganization(c.Service(), &form))
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
// @Success 501 {object} ctx.JSONResult{result=resps.UserWithRoleResp}
func (Organization) Delete(c *ctx.GinRequest) {
	form := forms.DeleteOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteOrganization(c.Service(), &form))
}

// Detail 组织信息详情
// @Tags 组织
// @Summary 组织信息详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param orgId path string true "组织ID"
// @router /orgs/{orgId} [get]
// @Success 200 {object} ctx.JSONResult{result=resps.OrganizationDetailResp}
func (Organization) Detail(c *ctx.GinRequest) {
	form := forms.DetailOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.OrganizationDetail(c.Service(), form))
}

// ChangeOrgStatus 启用/禁用组织
// @Tags 组织
// @Summary 启用/禁用组织
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param orgId path string false "组织ID"
// @Param form formData forms.DisableOrganizationForm true "parameter"
// @router /orgs/{orgId}/status [put]
// @Success 200 {object} ctx.JSONResult{result=models.Organization}
func (Organization) ChangeOrgStatus(c *ctx.GinRequest) {
	form := forms.DisableOrganizationForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ChangeOrgStatus(c.Service(), &form))
}

// AddUserToOrg 添加用户到组织
// @Tags 组织
// @Summary 添加用户到组织
// @Description 将一个或多个用户添加到组织里面。操作人需要拥有组织管理权限。
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param orgId path string true "组织ID"
// @Param form formData forms.AddUserOrgRelForm true "parameter"
// @router /orgs/{orgId}/users [post]
// @Success 200 {object} ctx.JSONResult{result=resps.UserWithRoleResp}
func (Organization) AddUserToOrg(c *ctx.GinRequest) {
	form := forms.AddUserOrgRelForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.AddUserOrgRel(c.Service(), &form))
}

// RemoveUserForOrg 从组织移除用户
// @Tags 组织
// @Summary 从组织移除用户
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param orgId path string true "组织ID"
// @Param userId path string true "用户ID"
// @Param form formData forms.DeleteUserOrgRelForm true "parameter"
// @router /orgs/{orgId}/users/{userId} [delete]
// @Success 200 {object} ctx.JSONResult{result=resps.UserWithRoleResp}
func (Organization) RemoveUserForOrg(c *ctx.GinRequest) {
	form := forms.DeleteUserOrgRelForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteUserOrgRel(c.Service(), &form))
}

// UpdateUserOrgRel 编辑用户组织角色
// @Tags 组织
// @Summary 编辑用户组织角色
// @Description 修改用户在组织中的角色。操作人需要拥有组织管理权限。
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param orgId path string true "组织ID"
// @Param userId path string true "用户ID"
// @Param form formData forms.UpdateUserOrgRelForm true "parameter"
// @router /orgs/{orgId}/users/{userId}/role [put]
// @Success 200 {object} ctx.JSONResult{result=resps.UserWithRoleResp}
func (Organization) UpdateUserOrgRel(c *ctx.GinRequest) {
	form := forms.UpdateUserOrgRelForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateUserOrgRel(c.Service(), &form))
}

// SearchUser 查询组织用户列表
// @Tags 组织
// @Summary 查询组织用户列表
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param orgId path string true "组织ID"
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.SearchUserForm true "parameter"
// @router /orgs/{orgId}/users [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.UserWithRoleResp}}
func (Organization) SearchUser(c *ctx.GinRequest) {
	form := forms.SearchUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.Service().OrgId = models.Id(c.Param("orgId"))
	c.JSONResult(apps.SearchUser(c.Service(), &form))
}

// InviteUser 邀请用户加入组织
// @Tags 组织
// @Summary 邀请内部或者外部用户加入组织
// @Description 如果用户不存在，则创建并加入组织，如果用户已经存在，则加入该组织
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form formData forms.InviteUserForm true "parameter"
// @Param orgId path string true "组织ID"
// @router /orgs/{orgId}/users/invite [post]
// @Success 200 {object} ctx.JSONResult{result=resps.UserWithRoleResp}
func (Organization) InviteUser(c *ctx.GinRequest) {
	form := forms.InviteUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.InviteUser(c.Service(), &form))
}

// SearchOrgResources 搜索当前组织下所有项目的活跃资源列表
// @Tags 组织
// @Summary 搜索当前组织下所有项目的活跃资源列表
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.SearchOrgResourceForm true "parameter"
// @router /orgs/resources [get]
// @Success 200 {object} ctx.JSONResult{result=resps.OrgOrProjectResourcesResp}
func (Organization) SearchOrgResources(c *ctx.GinRequest) {
	form := forms.SearchOrgResourceForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchOrgResources(c.Service(), &form))
}

// SearchOrgResourcesFilters 搜索当前组织下所有项目名称以及provider
// @Tags 组织
// @Summary 搜索当前组织下所有项目名称以及provider
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.SearchOrgResourceForm true "parameter"
// @router /orgs/resources/filters [get]
// @Success 200 {object} ctx.JSONResult{result=[]resps.OrgProjectAndProviderResp}
func (Organization) SearchOrgResourcesFilters(c *ctx.GinRequest) {
	form := forms.SearchOrgResourceForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchOrgResourcesFilters(c.Service(), &form))
}

// UpdateUserOrg 编辑组织用户信息
// @Tags 组织
// @Summary 编辑组织用户信息
// @Description 修改用户在组织中的角色。操作人需要拥有组织管理权限。
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param orgId path string true "组织ID"
// @Param userId path string true "用户ID"
// @Param form formData forms.UpdateUserOrgForm true "parameter"
// @router /orgs/{orgId}/users/{userId} [put]
// @Success 200 {object} ctx.JSONResult{result=resps.UserWithRoleResp}
func (Organization) UpdateUserOrg(c *ctx.GinRequest) {
	form := forms.UpdateUserOrgForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateUserOrg(c.Service(), &form))
}

// InviteUsersBatch 批量邀请用户加入组织
// @Tags 组织
// @Summary 邀请内部或者外部用户加入组织
// @Description 如果用户不存在，则创建并加入组织，如果用户已经存在，则加入该组织
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form formData forms.InviteUsersBatchForm true "parameter"
// @Param orgId path string true "组织ID"
// @router /orgs/{orgId}/users/batch_invite [post]
// @Success 200 {object} ctx.JSONResult{result=resps.InviteUsersBatchResp}
func (Organization) InviteUsersBatch(c *ctx.GinRequest) {
	form := forms.InviteUsersBatchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.InviteUsersBatch(c.Service(), &form))
}

// OrgProjectsStat 组织和项目概览统计数据
// @Tags 组织
// @Summary 组织和项目概览统计数据
// @Description 组织和项目概览统计数据
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form formData forms.OrgProjectsStatForm true "parameter"
// @router /orgs/projects/statistics [get]
// @Success 200 {object} ctx.JSONResult{result=resps.OrgProjectsStatResp}
func (Organization) OrgProjectsStat(c *ctx.GinRequest) {
	form := forms.OrgProjectsStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.OrgProjectsStat(c.Service(), &form))
}
