package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

// PortalSystemStatusSearch 查询系统状态列表
// @Summary 查询系统状态列表
// @Description 查询系统状态列表
// @Tags 系统状态
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} apps.SystemStatusResp
// @Router /system/status/search [get]
func PortalSystemStatusSearch(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.SystemStatusSearch())
}

// ConsulKVSearch 查询consul key value
// @Summary 查询consul key value
// @Description 查询consul key value
// @Tags 系统状态
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param key query string true "consul key"
// @Success 200 {object} string
// @Router /consul/kv/search [get]
func ConsulKVSearch(c *ctx.GinRequestCtx) {
	key := c.Query("key")
	c.JSONResult(apps.ConsulKVSearch(key))
}

// RunnerSearch 查询runner列表
// @Summary 查询runner列表
// @Description 查询runner列表
// @Tags 系统状态
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Success 200
// @Router /runner/search [get]
func RunnerSearch(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.RunnerSearch())
}

// ConsulTagUpdate 修改服务标签
// @Summary 修改服务标签
// @Description 修改服务标签
// @Tags 系统状态
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.ConsulTagUpdateForm true "tag信息"
// @Success 200
// @Router /consul/tags/update [put]
func ConsulTagUpdate(c *ctx.GinRequestCtx) {
	form:=forms.ConsulTagUpdateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ConsulTagUpdate(form))
}



