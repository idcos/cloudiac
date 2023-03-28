package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// GcpDeploy GCP环境创建
// @Tags 环境
// @Summary GCP环境创建
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param json body forms.GcpDeployForm true "parameter"
// @router /gcp/deploy [post]
// @Success 200 {object} ctx.JSONResult
func GcpDeploy(c *ctx.GinRequest) {
	// json字符串
	form := forms.GcpDeployForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.GcpDeploy(c.Service(), &form))
}
