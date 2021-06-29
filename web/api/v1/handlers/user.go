package handlers

import (
	"cloudiac/apps"
	"cloudiac/consts/e"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type User struct {
	ctrl.BaseController
}

// Create 创建用户
// @Summary 创建用户
// @Description 创建用户
// @Tags 用户
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.CreateUserForm true "用户信息"
// @Success 200 {object} models.User
// @Router /user/create [post]
func (User) Create(c *ctx.GinRequestCtx) {
	form := forms.CreateUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.CreateUser(c.ServiceCtx(), &form))
}

// Search 查询用户列表
// @Summary 查询用户列表
// @Description 查询用户列表
// @Tags 用户
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param q query string false "模糊搜索"
// @Param status query string false "用户状态"
// @Success 200 {object} models.User
// @Router /user/search [get]
func (User) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchUser(c.ServiceCtx(), &form))
}

// Update 修改用户信息
// @Summary 修改用户信息
// @Description 修改用户信息
// @Tags 用户
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.SearchUserForm true "用户信息"
// @Success 200 {object} models.User
// @Router /user/update [put]
func (User) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateUser(c.ServiceCtx(), &form))
}

func (User) Delete(c *ctx.GinRequestCtx) {
	//form := forms.DeleteUserForm{}
	//if err := c.Bind(&form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.DeleteUser(c.ServiceCtx(), &form))
	c.JSONError(e.New(e.NotImplement))
}

// Detail 用户信息详情
// @Summary 用户信息详情
// @Description 用户信息详情
// @Tags 用户
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param id query int true "用户id"
// @Success 200 {object} models.User
// @Router /user/detail [get]
func (User) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UserDetail(c.ServiceCtx(), form.Id))
}

// RemoveUserForOrg 组织用户删除
// @Summary 组织用户删除
// @Description 组织用户删除
// @Tags 用户
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.DeleteUserForm true "组织用户删除"
// @Success 200 {object} models.User
// @Router /org/user/delete [put]
func (User) RemoveUserForOrg(c *ctx.GinRequestCtx) {
	form := forms.DeleteUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteUserOrgMap(c.ServiceCtx(), &form))
}

// UserPassReset 修改用户密码
// @Summary 修改用户密码
// @Description 修改用户密码
// @Tags 用户
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.DetailUserForm true "修改用户密码"
// @Success 200 {object} models.User
// @Router /user/password/update [put]
func (User) UserPassReset(c *ctx.GinRequestCtx) {
	form := forms.DetailUserForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UserPassReset(c.ServiceCtx(), &form))
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录
// @Tags 用户
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.LoginForm true "用户登录"
// @Success 200 {object} models.User
// @Router /auth/login [post]
func (User) Login(c *ctx.GinRequestCtx) {
	form := forms.LoginForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.Login(c.ServiceCtx(), &form))
}

// GetUserByToken 登录用户信息
// @Summary 登录用户信息
// @Description 登录用户信息
// @Tags 用户
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} models.User
// @Router /user/info/search [get]
func (User) GetUserByToken(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.UserDetail(c.ServiceCtx(), c.ServiceCtx().UserId))
}
