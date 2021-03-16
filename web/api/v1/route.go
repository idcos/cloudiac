package v1

import (
	"github.com/gin-gonic/gin"
	"cloudiac/libs/ctrl"
	"cloudiac/web/middleware"
)

func Register(g *gin.RouterGroup) {
	w := ctrl.GinRequestCtxWrap

	/////// 用户认证
	g.Use(w(middleware.Auth))
	// g.Use(w(middleware.AlwaysAdminAuth))

	// 用户管理
	user := g.Group("/")
	//{
		ctrl.Register(user.Group("users"), &User{})
	//}

	g.Use(w(middleware.ApiAuth))

	//monitor := g.Group("")
	//{
		//ctrl.Register(monitor.Group("datasource"), &handlers.DataSource{})
		//monitor.GET("/datasource/metric/search", w(handlers.DataSource{}.SearchMetric))
		//monitor.GET("/datasource/relation_field/search", w(handlers.DataSource{}.SearchRelationField))
	//}
}
