package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"github.com/gin-gonic/gin"
)

type User struct {
	ctrl.BaseController
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
// @Success 200 {object} ctx.JSONResult{result=apps.CreateUserResp}
func (User) Create(c *ctx.GinRequestCtx) {
	form := forms.CreateUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.CreateUser(c.ServiceCtx(), &form))
}

// Search 用户查询
// @Tags 用户
// @Summary 用户查询
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.SearchUserForm true "parameter"
// @router /users [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.User}}
func (User) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchUser(c.ServiceCtx(), &form))
}

// Update 用户编辑
// @Tags 用户
// @Summary 用户信息编辑
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param userId path string true "用户ID"
// @Param form formData forms.UpdateUserForm true "parameter"
// @router /users/{userId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.User}
func (User) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateUser(c.ServiceCtx(), &form))
}

// ChangeUserStatus 启用/禁用用户
// @Tags 用户
// @Summary 启用/禁用用户
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param userId path string true "用户ID"
// @Param form formData forms.DisableUserForm true "parameter"
// @router /users/{userId}/status [put]
// @Success 200 {object} ctx.JSONResult{result=models.User}
func (User) ChangeUserStatus(c *ctx.GinRequestCtx) {
	form := forms.DisableUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ChangeUserStatus(c.ServiceCtx(), &form))
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
func (u User) UpdateSelf(c *ctx.GinRequestCtx) {
	// 将调用者 id 加入 Params 模拟 path 参数
	c.Params = append(c.Params, gin.Param{Key: "id", Value: string(c.ServiceCtx().UserId)})
	u.Update(c)
}

// Delete 删除用户
// @Tags 用户
// @Summary 删除用户
// @Description 需要组织管理员权限，如果用户拥有多个组织权限，管理员需要拥有所有相关组织权限。
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param userId path string true "用户ID"
// @Param form formData forms.DeleteUserForm true "parameter"
// @router /users/{userId} [delete]
// @Success 200 {object} ctx.JSONResult
func (User) Delete(c *ctx.GinRequestCtx) {
	form := forms.DeleteUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteUser(c.ServiceCtx(), &form))
}

// Detail 用户详情
// @Tags 用户
// @Summary 用户详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param userId path string true "用户ID"
// @router /users/{userId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.User}
func (User) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UserDetail(c.ServiceCtx(), form.Id))
}

// PasswordReset 重置用户密码
// @Tags 鉴权
// @Summary 用户重置密码
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Param IaC-Org-Id header string true "组织ID"
// @Param userId path string true "用户ID"
// @Param form formData forms.DetailUserForm true "parameter"
// @router /users/{userId}/reset_password [post]
// @Success 200 {object} ctx.JSONResult{result=apps.CreateUserResp}
func (User) PasswordReset(c *ctx.GinRequestCtx) {
	form := forms.DetailUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UserPassReset(c.ServiceCtx(), &form))
}
