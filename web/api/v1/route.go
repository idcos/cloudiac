package v1

import (
	"cloudiac/libs/ctrl"
	"cloudiac/web/api/v1/handlers"
	"cloudiac/web/middleware"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup) {
	w := ctrl.GinRequestCtxWrap

	auth := g.Group("/auth")
	auth.POST("/login", w(handlers.User{}.Login))

	/////// 用户认证
	// g.Use(w(middleware.Auth))
	// g.Use(w(middleware.AuthOrgId))
	user := g.Group("/")
	{
		ctrl.Register(user.Group("user"), &handlers.User{})
		user.PUT("/user/removeUserForOrg", w(handlers.User{}.RemoveUserForOrg))
		user.PUT("/user/userPassReset", w(middleware.IsOrgOwner), w(handlers.User{}.UserPassReset))

		ctrl.Register(user.Group("org"), &handlers.Organization{})
		user.PUT("/org/disableOrg", w(middleware.IsAdmin), w(handlers.Organization{}.DisableOrganization))
		//root.GET("/org/detail", w(handlers.Organization{}.Detail))
	}

	user.GET("/sse/hello/:filename", w(handlers.HelloSse))
	user.GET("/sse/test", w(handlers.TestSSE))

	//g.Use(w(middleware.ApiAuth))

	//monitor := g.Group("")
	//{
	//ctrl.Register(monitor.Group("datasource"), &handlers.DataSource{})
	//monitor.GET("/datasource/metric/search", w(handlers.DataSource{}.SearchMetric))
	//monitor.GET("/datasource/relation_field/search", w(handlers.DataSource{}.SearchRelationField))
	//}

	// 系统状态
	g.GET("/systemStatus/search", w(handlers.PortalSystemStatusSearch))
}
