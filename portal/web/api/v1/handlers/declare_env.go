package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"net/http"
)

// DeclareEnv CPG环境创建
// @Tags 环境
// @Summary CPG环境创建
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param json body forms.DeclareEnvForm true "parameter"
// @router /declare/env [post]
// @Success 200 {object} ctx.JSONResult
func DeclareEnv(c *ctx.GinRequest) {
	// 鉴权
	token := c.GetHeader("Authorization")
	if token != "34430039-6c5d-4ff1-8185-0c66eb7738a1" {
		c.JSONError(e.New(e.InvalidToken), http.StatusUnauthorized)
		return
	}

	form := forms.DeclareEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeclareEnv(c.Service(), &form))
}
