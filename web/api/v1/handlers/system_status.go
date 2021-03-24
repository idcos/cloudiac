package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctx"
)

func PortalSystemStatusSearch(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.SystemStatusSearch())
}


