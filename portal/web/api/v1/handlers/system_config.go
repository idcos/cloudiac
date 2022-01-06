// Copyright 2021 CloudJ Company Limited. All rights reserved.

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
// @Success 200 {object} ctx.JSONResult{result=[]apps.SearchSystemConfigResp}
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
// @Param id path string true "系统ID"
// @Param data body forms.UpdateSystemConfigForm true "系统配置信息"
// @Success 200 {object} models.SystemCfg
// @Router /systems [put]
func (SystemConfig) Update(c *ctx.GinRequest) {
	form := forms.UpdateSystemConfigForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateSystemConfig(c.Service(), &form))
}

// CheckRegistryAddr 检查registry addr是否已配置
// @Summary 检查registry addr是否已配置(先检查DB，再检查配置文件)
// @Tags registry
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Success 200 {object}  ctx.JSONResult{result=models.RegistryAddrCheckResp}
// @Router /sys/registry/check [GET]
func CheckRegistryAddr(c *ctx.GinRequest) {
	c.JSONResult(apps.CheckRegistryAddr(c.Service()))
}
