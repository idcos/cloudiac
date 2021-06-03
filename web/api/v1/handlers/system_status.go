package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

func PortalSystemStatusSearch(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.SystemStatusSearch())
}

func ConsulKVSearch(c *ctx.GinRequestCtx) {
	key := c.Query("key")
	c.JSONResult(apps.ConsulKVSearch(key))
}

func RunnerSearch(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.RunnerSearch())
}

func ConsulTagUpdate(c *ctx.GinRequestCtx) {
	form:=forms.ConsulTagUpdateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ConsulTagUpdate(form))
}



