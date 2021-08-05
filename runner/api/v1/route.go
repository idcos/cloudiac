// Copyright 2021 CloudJ Company Limited. All rights reserved.

package v1

import (
	"cloudiac/common"
	"cloudiac/runner/api/ctx"
	"cloudiac/runner/api/v1/handler"
	"github.com/gin-gonic/gin"
)

func RegisterRoute(apiV1 *gin.RouterGroup) {
	w := ctx.HandlerWrapper

	apiV1.Any("/check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
			"version": common.VERSION,
			"build":   common.BUILD,
		})
	})

	apiV1.Use(gin.Logger())
	apiV1.POST("/task/run", w(handler.RunTask))
	apiV1.GET("/task/status", w(handler.TaskStatus))
	apiV1.GET("/task/log/follow", w(handler.TaskLogFollow))
}
