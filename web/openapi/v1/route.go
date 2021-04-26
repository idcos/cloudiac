package v1

import (
	"cloudiac/libs/ctrl"
	"cloudiac/web/middleware"
	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup) {
	w := ctrl.GinRequestCtxWrap

	iac := g.Group("/iac", w(middleware.OpenApiAuth))
	iac.GET("/")

}
