package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
)

// GcpDeploy 获取任务资源列表
// @Tags 环境
// @Summary 获取任务资源列表
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param form query forms.SearchTaskResourceGraphForm true "parameter"
// @Param taskId path string true "任务ID"
// @router /tasks/{taskId}/resources/graph [get]
// @Success 200 {object} ctx.JSONResult
func GcpDeploy(c *ctx.GinRequest) {
	// json字符串

	c.JSONResult(apps.GcpDeploy(c.Service()))
}
