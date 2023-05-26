package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
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
	form := forms.DeclareEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeclareEnv(c.Service(), &form))
}
