package v1

import (
	"github.com/gin-gonic/gin"

	"cloudiac/libs/ctrl"
	api_handlers "cloudiac/web/api/v1/handlers"
	"cloudiac/web/middleware"
	"cloudiac/web/openapi/v1/handlers"
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
	g.GET("/taskLog/sse", w(api_handlers.Task{}.FollowLogSse))
	iac.POST("/task/create", w(handlers.TaskCreate))

}
