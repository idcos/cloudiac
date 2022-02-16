// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

func WebhooksApiHandler(c *ctx.GinRequest) {
	form := forms.WebhooksApiHandler{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.WebhooksApiHandler(c.Service(), form))
}
