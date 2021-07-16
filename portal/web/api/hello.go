package api

import (
	"cloudiac/common"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"

	"cloudiac/portal/libs/ctx"
)

// 简单 handler 函数
func Hello(c *ctx.GinRequestCtx) {
	c.JSON(http.StatusOK, "", gin.H{
		"hello": "world",
		"goos":  runtime.GOOS,
	})
}

func Health(c *ctx.GinRequestCtx) {
	c.JSON(http.StatusOK, nil, gin.H{
		"ok":      true,
		"version": common.VERSION,
		"build":   common.BUILD,
	})
}
