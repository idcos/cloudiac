package v1

import (
	"cloudiac/libs/ctrl"
	"cloudiac/web/middleware"
	"cloudiac/web/openapi/v1/handlers"
	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup) {
	w := ctrl.GinRequestCtxWrap
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
	iac.GET("/template/detail", w(handlers.TemplateDetail))
	iac.GET("/runnerList/search", w(handlers.RunnerListSearch))
	iac.GET("/template/search", w(handlers.OpenTemplateSearch))
	g.GET("/taskLog/sse", w(handlers.TaskLogSSE))
	iac.POST("/task/create", w(handlers.TaskCreate))

}
