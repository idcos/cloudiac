// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// WebhooksApiHandler Webhooks api处理器
// @Tags Webhooks
// @Summary Webhooks api处理器
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Param vcsType path string true "vcs 类型"
// @Param vcsId path string true "vcs ID"
// @Param json body forms.WebhooksApiHandler true "parameter"
// @Router /webhooks/{vcsType}/{vcsId} [post]
// @Success 200 {object} ctx.JSONResult
func WebhooksApiHandler(c *ctx.GinRequest) {
	form := forms.WebhooksApiHandler{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.WebhooksApiHandler(c.Service(), form))
}
