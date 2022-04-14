// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// GenerateSsoToken 生成 SSO token
// @Summary 生成 token 供单点登录使用
// @Tags SSO
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Success 200 {object}  ctx.JSONResult{result=resps.SsoResp}
// @Router /sso/tokens [post]
func GenerateSsoToken(c *ctx.GinRequest) {
	c.JSONResult(apps.GenerateSsoToken(c.Service()))
}

// VerifySsoToken 验证 SSO token
// @Summary 验证单点登录的 token，返回单点登录的用户信息
// @Tags SSO
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param token query string true "sso token"
// @Success 200 {object}  ctx.JSONResult{result=resps.VerifySsoTokenResp}
// @Router /sso/tokens/verify [get]
func VerifySsoToken(c *ctx.GinRequest) {
	form := forms.VerifySsoTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}

	c.JSONResult(apps.VerifySsoToken(c.Service(), &form))
}
