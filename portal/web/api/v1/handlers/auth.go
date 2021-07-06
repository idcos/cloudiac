package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Auth struct {
	ctrl.BaseController
}

func (Auth) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateToken(c.ServiceCtx(), form))
}

func (Auth) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchToken(c.ServiceCtx(), form))
}

func (Auth) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateToken(c.ServiceCtx(), form))
}

func (Auth) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteToken(c.ServiceCtx(), form))
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
func (Auth) GetUserByToken(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.UserDetail(c.ServiceCtx(), c.ServiceCtx().UserId))
}

// Login 用户登陆
// @Tags 鉴权
// @Summary 用户登陆
// @Accept multipart/form-data
// @Accept json
// @Param body formData forms.LoginForm true "parameter"
// @router /auth/login [post]
// @Success 200 {object} ctx.JSONResult{result=models.LoginResp}
func (Auth) Login(c *ctx.GinRequestCtx) {
	form := forms.LoginForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.Login(c.ServiceCtx(), &form))
}
