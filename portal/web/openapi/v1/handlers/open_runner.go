package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
)

func RunnerListSearch(c *ctx.GinRequest) {
	c.JSONResult(apps.RunnerSearch())
}
