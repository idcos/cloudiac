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
// @Success 200 {object}  ctx.JSONResult{result=models.SsoResp}
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
// @Success 200 {object}  ctx.JSONResult{result=models.VerifySsoTokenResp}
// @Router /sso/tokens [post]
func VerifySsoToken(c *ctx.GinRequest) {
	form := forms.VerifySsoTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}

	c.JSONResult(apps.VerifySsoToken(c.Service(), &form))
}
