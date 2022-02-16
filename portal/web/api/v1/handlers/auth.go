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
// @Success 200 {object} ctx.JSONResult{result=models.User}
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
// @Success 200 {object} ctx.JSONResult{result=models.LoginResp}
func (a Auth) Login(c *ctx.GinRequest) {
	form := forms.LoginForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.Login(c.Service(), &form))
}
