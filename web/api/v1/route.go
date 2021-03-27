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

	g.GET("/org/search", w(middleware.Auth), w(handlers.Organization{}.Search))
	g.GET("/org/detail", w(middleware.Auth), w(handlers.Organization{}.Detail))
	org := g.Group("/", w(middleware.Auth), w(middleware.IsAdmin))
	{
		org.POST("/org/create", w(handlers.Organization{}.Create))
		org.PUT("/org/update", w(handlers.Organization{}.Update))
		org.PUT("/org/changeStatus", w(handlers.Organization{}.ChangeOrgStatus))
	}

	user := g.Group("/", w(middleware.Auth), w(middleware.AuthOrgId))

	{
		user.GET("/user/search", w(middleware.IsOrgOwner), w(handlers.User{}.Search))
		user.GET("/user/detail", w(middleware.IsOrgOwner), w(handlers.User{}.Detail))
		user.POST("/user/create", w(middleware.IsOrgOwner), w(handlers.User{}.Create))
		user.PUT("/user/update", w(handlers.User{}.Update))
		user.PUT("/user/removeUserForOrg", w(middleware.IsOrgOwner), w(handlers.User{}.RemoveUserForOrg))
		user.PUT("/user/userPassReset", w(middleware.IsOrgOwner), w(handlers.User{}.UserPassReset))

		user.GET("/org/listRepos", w(handlers.GitLab{}.ListRepos))
		user.GET("/org/listBranches", w(handlers.GitLab{}.ListBranches))
		user.GET("/org/getReadme", w(handlers.GitLab{}.GetReadmeContent))

		user.GET("/org/notification/search", w(handlers.Organization{}.ListNotificationCfgs))
		user.POST("/org/notification/create", w(handlers.Organization{}.CreateNotificationCfgs))
		user.DELETE("/org/notification/delete", w(handlers.Organization{}.DeleteNotificationCfgs))
		user.PUT("/org/notification/update", w(handlers.Organization{}.UpdateNotificationCfgs))
		//root.GET("/org/detail", w(handlers.Organization{}.Detail))
	}
	sysConf := g.Group("/", w(middleware.Auth), w(middleware.IsAdmin))
	{
		sysConf.GET("/system/search", w(handlers.SystemConfig{}.Search))
		sysConf.PUT("/system/update", w(handlers.SystemConfig{}.Update))
	}

	template := g.Group("/", w(middleware.Auth), w(middleware.IsAdmin))
	{
		ctrl.Register(template.Group("template"), &handlers.Template{})
	}

	user.GET("/sse/hello/:filename", w(handlers.HelloSse))
	user.GET("/sse/test", w(handlers.TestSSE))

	// 系统状态
	g.GET("/systemStatus/search", w(handlers.PortalSystemStatusSearch))

	//实现task create接口
	//实现task排队调度过程
	//实现task detail接口
	//实现task logs接口
	task := g.Group("/task")
	{
		task.GET("/detail",w(handlers.TaskDetail))
		task.POST("/create",w(handlers.TaskCreate))
	}
}
