package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
)

// DriftDetail 漂移信息
// @Tags 环境
// @Summary 环境重新部署漂移检测
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/drift/detail
// @Success 200 {object} ctx.JSONResult{}
func DriftDetail(c *ctx.GinRequest) {
	c.JSONResult(apps.EnvDriftDetail(c.Service(), models.Id(c.Param("id"))))
}
