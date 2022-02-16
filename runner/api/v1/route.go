// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

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
	apiV1.POST("/task/step/run", w(handler.RunTask))
	apiV1.GET("/task/step/status", w(handler.TaskStatus))
	apiV1.POST("/task/stop", w(handler.StopTask))
	apiV1.GET("/task/step/log/follow", w(handler.TaskLogFollow))
}
