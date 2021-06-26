package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type SystemConfig struct {
	ctrl.BaseController
}

// Create 创建系统配置
// @Summary 创建系统配置
// @Description 创建系统配置
// @Tags 系统配置
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.CreateOrganizationForm true "系统配置信息"
// @Success 200 {object} models.Organization
// @Router /system/create [post]
func (SystemConfig) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateOrganizationForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateOrganization(c.ServiceCtx(), form))
}

// Search 查询系统配置列表
// @Summary 查询系统配置列表
// @Description 查询系统配置列表
// @Tags 系统配置
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} apps.searchSystemConfigResp
// @Router /system/search [get]
func (SystemConfig) Search(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.SearchSystemConfig(c.ServiceCtx()))
}

// Update 修改系统配置信息
// @Summary 修改系统配置信息
// @Description 修改系统配置信息
// @Tags 系统配置
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.UpdateSystemConfigForm true "系统配置信息"
// @Success 200 {object} models.SystemCfg
// @Router /system/update [put]
func (SystemConfig) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateSystemConfigForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateSystemConfig(c.ServiceCtx(), &form))
}
