package v1

import (
	"cloudiac/libs/ctrl"
	"cloudiac/web/api/v1/handlers"
	"cloudiac/web/middleware"

	"github.com/gin-gonic/gin"
)

func Register(g *gin.RouterGroup) {
	w := ctrl.GinRequestCtxWrap

	g.Any("/check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
		})
	})
	auth := g.Group("/auth")
	auth.POST("/login", w(handlers.User{}.Login))

	g.GET("/org/search", w(middleware.Auth), w(handlers.Organization{}.Search))
	g.GET("/org/detail", w(middleware.Auth), w(handlers.Organization{}.Detail))

	// IaC管理员权限
	sys := g.Group("/", w(middleware.Auth), w(middleware.IsAdmin))
	{
		sys.POST("/org/create", w(handlers.Organization{}.Create))
		sys.PUT("/org/update", w(handlers.Organization{}.Update))
		sys.PUT("/org/changeStatus", w(handlers.Organization{}.ChangeOrgStatus))

		ctrl.Register(sys.Group("system"), &handlers.SystemConfig{})
		ctrl.Register(sys.Group("token"), &handlers.Token{})
	}

	root := g.Group("/", w(middleware.Auth), w(middleware.AuthOrgId))
	{
		root.GET("/user/search", w(middleware.IsOrgOwner), w(handlers.User{}.Search))
		root.GET("/user/detail", w(middleware.IsOrgOwner), w(handlers.User{}.Detail))
		root.POST("/user/create", w(middleware.IsOrgOwner), w(handlers.User{}.Create))
		root.PUT("/user/update", w(handlers.User{}.Update))
		root.PUT("/user/removeUserForOrg", w(middleware.IsOrgOwner), w(handlers.User{}.RemoveUserForOrg))
		root.PUT("/user/userPassReset", w(middleware.IsOrgOwner), w(handlers.User{}.UserPassReset))

		root.GET("/gitlab/listRepos", w(handlers.GitLab{}.ListRepos))
		root.GET("/gitlab/listBranches", w(handlers.GitLab{}.ListBranches))
		root.GET("/gitlab/getReadme", w(handlers.GitLab{}.GetReadmeContent))

		ctrl.Register(root.Group("notification"), &handlers.Notification{})
		ctrl.Register(root.Group("resourceAccount"), &handlers.ResourceAccount{})
		ctrl.Register(root.Group("template"), &handlers.Template{})
		ctrl.Register(root.Group("task"), &handlers.Task{})
	}

	root.GET("/sse/hello/:filename", w(handlers.HelloSse))
	root.GET("/sse/test", w(handlers.TestSSE))
	root.GET("/task_log/sse", w(handlers.TaskLogSSE))

	// 系统状态
	g.GET("/systemStatus/search", w(handlers.PortalSystemStatusSearch))

}
