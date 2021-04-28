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
	iac.GET("/template/search", w(handlers.OpenTemplateSearch))
	g.GET("/taskLog/sse", w(handlers.TaskLogSSE))
	//iac.GET("/task/create",w(handlers2.TaskCreate))
}
