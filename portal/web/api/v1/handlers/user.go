// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"github.com/gin-gonic/gin"
)

type User struct {
	ctrl.GinController
}

// Create 创建用户
// @Tags 用户
// @Summary 创建用户
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form formData forms.CreateUserForm true "parameter"
// @router /users [post]
// @Success 200 {object} ctx.JSONResult{result=resps.CreateUserResp}
func (User) Create(c *ctx.GinRequest) {
	form := forms.CreateUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.CreateUser(c.Service(), &form))
}

// Search 用户列表查询
// @Tags 用户
// @Summary 用户列表查询
// @Description 平台管理员可以查询所有用户，不附带 IaC-Org-Id 和 IaC-Project-Id header
// @Description 组织内用户可以查询本组织用户列表，附带 IaC-Org-Id header
// @Description 项目内用户可以查询本项目用户列表，附带 IaC-Org-Id 和 IaC-Project-Id header
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string false "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.SearchUserForm true "parameter"
// @router /users [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.UserWithRoleResp}}
func (User) Search(c *ctx.GinRequest) {
	form := forms.SearchUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchUser(c.Service(), &form))
}

// Update 用户编辑
// @Tags 用户
// @Summary 用户信息编辑
// @Description 用户可以编辑自己，组织管理员可以编辑组织下的用户，平台管理员可以编辑所有用户
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string false "组织ID"
// @Param userId path string true "用户ID"
// @Param form formData forms.UpdateUserForm true "parameter"
// @router /users/{userId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.User}
func (User) Update(c *ctx.GinRequest) {
	form := forms.UpdateUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateUser(c.Service(), &form))
}

// ChangeUserStatus 启用/禁用用户
// @Tags 用户
// @Summary 启用/禁用用户
// @Description 需要平台管理员权限
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string false "组织ID"
// @Param userId path string true "用户ID"
// @Param form formData forms.DisableUserForm true "parameter"
// @router /users/{userId}/status [put]
// @Success 200 {object} ctx.JSONResult{result=models.User}
func (User) ChangeUserStatus(c *ctx.GinRequest) {
	form := forms.DisableUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ChangeUserStatus(c.Service(), &form))
}

// UpdateSelf 用户自身信息编辑
// @Tags 用户
// @Summary 用户自身信息编辑
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.UpdateUserForm true "parameter"
// @router /users/self [put]
// @Success 200 {object} ctx.JSONResult{result=models.User}
func (u User) UpdateSelf(c *ctx.GinRequest) {
	// 将调用者 id 加入 Params 模拟 path 参数
	c.Params = append(c.Params, gin.Param{Key: "id", Value: string(c.Service().UserId)})
	u.Update(c)
}

// Delete 删除用户
// @Tags 用户
// @Summary 删除用户
// @Description 需要平台管理员权限
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string false "组织ID"
// @Param userId path string true "用户ID"
// @router /users/{userId} [delete]
// @Success 200 {object} ctx.JSONResult
func (User) Delete(c *ctx.GinRequest) {
	form := forms.DeleteUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteUser(c.Service(), &form))
}

// Detail 用户详情
// @Tags 用户
// @Summary 用户详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string false "组织ID"
// @Param userId path string true "用户ID"
// @router /users/{userId} [get]
// @Success 200 {object} ctx.JSONResult{result=resps.UserWithRoleResp}
func (User) Detail(c *ctx.GinRequest) {
	form := forms.DetailUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UserDetail(c.Service(), form.Id))
}

// PasswordReset 重置用户密码
// @Tags 用户
// @Summary 用户重置密码
// @Description 需要平台管理员或者组织管理员权限
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Param IaC-Org-Id header string true "组织ID"
// @Param userId path string true "用户ID"
// @router /users/{userId}/password/reset [post]
// @Success 200 {object} ctx.JSONResult{result=models.User}
func (User) PasswordReset(c *ctx.GinRequest) {
	form := forms.DetailUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UserPassReset(c.Service(), &form))
}

// LdapSearch 平台所有用户查询
// @Tags 用户
// @Summary 用户列表查询
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param form query forms.SearchUserForm true "parameter"
// @router /users/all [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.UserWithRoleResp}}
func (User) SearchAllUsers(c *ctx.GinRequest) {
	form := forms.SearchUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchAllUser(c.Service(), &form))
}

// ActiveUserEmail 邮箱激活
// @Tags 邮箱
// @Summary 邮箱激活
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @router /register/activation/ [get]
// @Success 200 {object} ctx.JSONResult{result=models.User}
func (User) ActiveUserEmail(c *ctx.GinRequest) {
	c.JSONResult(apps.ActiveUserEmail(c.Service(), c.Service().UserId))
}
