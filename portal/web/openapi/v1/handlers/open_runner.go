package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
)

func RunnerListSearch(c *ctx.GinRequestCtx) {
	c.JSONOpenResultList(apps.RunnerSearch())
}
