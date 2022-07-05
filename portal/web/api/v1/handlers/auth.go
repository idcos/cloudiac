// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Auth struct {
	ctrl.GinController
}

func (Auth) Create(c *ctx.GinRequest) {
	form := &forms.CreateTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateToken(c.Service(), form))
}

func (Auth) Search(c *ctx.GinRequest) {
	form := &forms.SearchTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchToken(c.Service(), form))
}

func (Auth) Update(c *ctx.GinRequest) {
	form := &forms.UpdateTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateToken(c.Service(), form))
}

func (Auth) Delete(c *ctx.GinRequest) {
	form := &forms.DeleteTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteToken(c.Service(), form))
}

// GetUserByToken 获取登陆用户自身信息
// @Tags 鉴权
// @Summary 获取登陆用户自身信息
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @router /auth/me [get]
// @Success 200 {object} ctx.JSONResult{result=resps.UserWithRoleResp}
func (Auth) GetUserByToken(c *ctx.GinRequest) {
	c.JSONResult(apps.UserDetail(c.Service(), c.Service().UserId))
}

// Login 用户登陆
// @Tags 鉴权
// @Summary 用户登陆
// @Accept multipart/form-data
// @Accept json
// @Param body formData forms.LoginForm true "parameter"
// @router /auth/login [post]
// @Success 200 {object} ctx.JSONResult{result=resps.LoginResp}
func (a Auth) Login(c *ctx.GinRequest) {
	form := forms.LoginForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.Login(c.Service(), &form))
}

// Registry 账号注册
// @Tags 鉴权
// @Summary 账号注册
// @Accept multipart/form-data
// @Accept json
// @Param body formData forms.RegistryForm true "parameter"
// @router /auth/register [post]
// @Success 200 {object} ctx.JSONResult{}
func (a Auth) Registry(c *ctx.GinRequest) {
	form := forms.RegistryForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.Register(c.Service(), &form))
}

// CheckEmail 用户邮箱验重
// @Tags 验证
// @Summary 用户邮箱验重
// @Accept multipart/form-data
// @Accept json
// @Param body formData forms.EmailForm true "parameter"
// @router /auth/email [get]
// @Success 200 {object} ctx.JSONResult{result=resps.UserEmailStatus}
func (a Auth) CheckEmail(c *ctx.GinRequest) {
	form := forms.EmailForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.CheckEmail(c.Service(), &form))
}

// PasswordResetEmail 发送找回密码邮件
// @Tags 验证
// @Summary 发送找回密码邮件
// @Accept multipart/form-data
// @Accept json
// @Param body formData forms.PasswordResetEmailForm true "parameter"
// @router /auth/password/reset [put]
// @Success 200 {object} ctx.JSONResult{result=resps.UserEmailStatus}
func (a Auth) PasswordResetEmail(c *ctx.GinRequest) {
	form := forms.PasswordResetEmailForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PasswordResetToSendEmail(c.Service(), &form))
}

// PasswordReset 找回密码
// @Tags 验证
// @Summary 找回密码
// @Accept multipart/form-data
// @Accept json
// @Param body formData forms.PasswordResetForm true "parameter"
// @router /auth/password/reset [put]
// @Success 200 {object} ctx.JSONResult{result=resps.UserEmailStatus}
func (a Auth) PasswordReset(c *ctx.GinRequest) {
	form := forms.PasswordResetForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PasswordReset(c.Service(), &form))
}
