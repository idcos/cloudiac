package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
)

// DriftDetail 漂移信息
// @Tags 环境
// @Summary 环境重新部署漂移检测
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/drift/detail [get]
// @Success 200 {object} ctx.JSONResult{result=models.EnvDrift}
func DriftDetail(c *ctx.GinRequest) {
	c.JSONResult(apps.EnvDriftDetail(c.Service(), models.Id(c.Param("id"))))
}

// DriftList 漂移信息列表
// @Tags 环境
// @Summary 环境漂移信息列表
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param form query forms.SearchEnvDriftsForm true "parameter"
// @router /envs/{envId}/drift [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.TaskDriftResp}}
func DriftList(c *ctx.GinRequest) {
	form := &forms.SearchEnvDriftsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.EnvDriftSearch(c.Service(), models.Id(c.Param("id")), form))
}

// DriftResourceList 漂移资源列表
// @Tags 环境
// @Summary 环境漂移信息列表
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param taskId path string true "漂移检查任务ID"
// @router /envs/{envId}/drift/{taskId}/resources [get]
// @Success 200 {object} ctx.JSONResult{result=resps.ResourceDriftResp}
func DriftResourceList(c *ctx.GinRequest) {
	c.JSONResult(apps.EnvDriftResourceSearch(c.Service(), models.Id(c.Param("id")), models.Id(c.Param("taskId"))))
}
