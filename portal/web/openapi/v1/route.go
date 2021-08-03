// Copyright 2021 CloudJ Company Limited. All rights reserved.

package v1

import (
	"github.com/gin-gonic/gin"

	"cloudiac/portal/libs/ctrl"
	api_handlers "cloudiac/portal/web/api/v1/handlers"
	"cloudiac/portal/web/middleware"
	"cloudiac/portal/web/openapi/v1/handlers"
)

func Register(g *gin.RouterGroup) {
	w := ctrl.WrapHandler
	iac := g.Group("/", w(middleware.OpenApiAuth))
	g.Any("/check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
		})
	})
	iac.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
		})
	})
	//iac.GET("/template/detail", w(handlers.TemplateDetail))
	iac.GET("/runnerList/search", w(handlers.RunnerListSearch))
	//iac.GET("/template/search", w(handlers.OpenTemplateSearch))
	g.GET("/taskLog/sse", w(api_handlers.Task{}.FollowLogSse))
	//iac.POST("/task/create", w(handlers.TaskCreate))

	iac.GET("/template/detail", w(handlers.TemplateDetail))
	iac.GET("/template/search", w(handlers.OpenTemplateSearch))
	iac.GET("/runner/search", w(handlers.RunnerListSearch))
	iac.POST("/task/create", w(handlers.TaskCreate))
	g.GET("/task/log/sse", w(api_handlers.Task{}.FollowLogSse))

}
