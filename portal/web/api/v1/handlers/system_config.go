package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type SystemConfig struct {
	ctrl.BaseController
}

func (SystemConfig) Create(c *ctx.GinRequestCtx) {
	//form := &forms.CreateOrganizationForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.CreateOrganization(c.ServiceCtx(), form))
}

// Search 查询系统配置列表
// @Summary 查询系统配置列表
// @Description 查询系统配置列表
// @Tags 系统配置
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Success 200 {object} ctx.JSONResult{result=apps.SearchSystemConfigResp}
// @Router /systems [get]
func (SystemConfig) Search(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.SearchSystemConfig(c.ServiceCtx()))
}

// Update 修改系统配置信息
// @Summary 修改系统配置信息
// @Description 修改系统配置信息
// @Tags 系统配置
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param id path string true "系统ID"
// @Param data body forms.UpdateSystemConfigForm true "系统配置信息"
// @Success 200 {object} models.SystemCfg
// @Router /systems/{id} [put]
func (SystemConfig) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateSystemConfigForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateSystemConfig(c.ServiceCtx(), &form))
}
