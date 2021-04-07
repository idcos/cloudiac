package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctx"
)

func PortalSystemStatusSearch(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.SystemStatusSearch())
}

func ConsulKVSearch(c *ctx.GinRequestCtx) {
	key := c.Query("key")
	c.JSONResult(apps.ConsulKVSearch(key))
}

func RunnerListSearch(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.RunnerListSearch())
}



