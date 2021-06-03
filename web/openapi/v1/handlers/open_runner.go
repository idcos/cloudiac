package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctx"
)


func RunnerListSearch(c *ctx.GinRequestCtx) {
	c.JSONOpenResultList(apps.RunnerSearch())
}

