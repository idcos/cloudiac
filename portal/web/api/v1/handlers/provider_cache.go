// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// SystemProviderCacheClear Delete 清理 provider 缓存
// @Summary 清理 provider 缓存
// @Description 清理 provider 缓存
// @Tags provider 缓存
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param json body forms.ClearProviderCacheForm true "parameter"
// @Success 200 {object}  ctx.JSONResult
// @Router /systems/provider_cache/remove [post]
func SystemProviderCacheClear(c *ctx.GinRequest) {
	form := &forms.ClearProviderCacheForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.ClearProviderCache(c.Service(), form))
}
