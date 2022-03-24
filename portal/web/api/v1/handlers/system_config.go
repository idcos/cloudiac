// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type SystemConfig struct {
	ctrl.GinController
}

func (SystemConfig) Create(c *ctx.GinRequest) {
	//form := &forms.CreateOrganizationForm{}
	//if err := c.IsBind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.CreateOrganization(c.ServiceContext(), form))
}

// Search 查询系统配置列表
// @Summary 查询系统配置列表
// @Description 查询系统配置列表
// @Tags 系统配置
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Success 200 {object} ctx.JSONResult{result=[]resps.SearchSystemConfigResp}
// @Router /systems [get]
func (SystemConfig) Search(c *ctx.GinRequest) {
	c.JSONResult(apps.SearchSystemConfig(c.Service()))
}

// Update 修改系统配置信息
// @Summary 修改系统配置信息
// @Description 修改系统配置信息
// @Tags 系统配置
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param data body forms.UpdateSystemConfigForm true "系统配置信息"
// @Success 200 {object} ctx.JSONResult{result=models.SystemCfg}
// @Router /systems [put]
func (SystemConfig) Update(c *ctx.GinRequest) {
	form := forms.UpdateSystemConfigForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateSystemConfig(c.Service(), &form))
}

// GetRegistryAddr 获取 registry addr 的配置
// @Summary 获取 registry addr 的配置
// @Tags registry
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Success 200 {object}  ctx.JSONResult{result=resps.RegistryAddrResp}
// @Router /system_config/registry/addr [GET]
func GetRegistryAddr(c *ctx.GinRequest) {
	c.JSONResult(apps.GetRegistryAddr(c.Service()))
}

// UpsertRegistryAddr 更新或创建 registry addr 的配置
// @Summary 更新或创建 registry addr 的配置
// @Tags registry
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param data body forms.RegistryAddrForm true "系统配置信息"
// @Success 200 {object}  ctx.JSONResult{result=resps.RegistryAddrResp}
// @Router /system_config/registry/addr [POST]
func UpsertRegistryAddr(c *ctx.GinRequest) {
	form := forms.RegistryAddrForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpsertRegistryAddr(c.Service(), &form))
}
