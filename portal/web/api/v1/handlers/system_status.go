// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// PortalSystemStatusSearch 查询系统状态
// @Summary 查询系统状态
// @Description 查询系统状态
// @Tags 系统状态
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Success 200 {object} ctx.JSONResult{list=[]resps.SystemStatusResp}
// @Router /systems/status [get]
func PortalSystemStatusSearch(c *ctx.GinRequest) {
	c.JSONResult(apps.SystemStatusSearch())
}

// SystemSwitchesStatus 查询系统功能开关状态
// @Summary 查询系统功能开关状态
// @Description 查询系统功能开关状态
// @Tags 查询系统功能开关状态
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Success 200 {object} ctx.JSONResult{abortStatus=bool}
// @Router /system_config/switches [get]
func SystemSwitchesStatus(c *ctx.GinRequest) {
	c.JSONResult(apps.SystemSwitchStatus())
}

// ConsulKVSearch 服务查询
// @Summary 查询系统状态
// @Description 查询系统状态
// @Tags 系统状态
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param key query string true "key"
// @Success 200 {object} ctx.JSONResult{result=string}
// @Router /consul/kv/search [get]
func ConsulKVSearch(c *ctx.GinRequest) {
	key := c.Query("key")
	c.JSONResult(apps.ConsulKVSearch(key))
}

// RunnerSearch 查询runner列表
// @Summary 查询runner列表
// @Description 查询runner列表
// @Tags runner
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Success 200 {object} ctx.JSONResult
// @Router /runners [get]
func RunnerSearch(c *ctx.GinRequest) {
	c.JSONResult(apps.RunnerSearch())
}

// ConsulTagUpdate 修改服务标签
// @Summary 修改服务标签
// @Description 修改服务标签
// @Tags 系统状态
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param data body forms.ConsulTagUpdateForm true "tag信息"
// @Success 200 {object} ctx.JSONResult
// @Router /consul/tags/update [put]
func ConsulTagUpdate(c *ctx.GinRequest) {
	form := forms.ConsulTagUpdateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ConsulTagUpdate(c.Service(), form))
}

// RunnerTags 查询runner tags列表
// @Summary 查询runner tags列表
// @Description 查询runner tags列表
// @Tags runner
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Success 200 {object} ctx.JSONResult{result=resps.RunnerTagsResp}
// @Router /runners/tags [get]
func RunnerTags(c *ctx.GinRequest) {
	c.JSONResult(apps.RunnerTags())
}
