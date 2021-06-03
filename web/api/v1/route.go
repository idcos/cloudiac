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

	{
		o := g.Group("/", w(middleware.Auth))
		//o.GET("/org/search", w(handlers.Organization{}.Search))
		//o.GET("/org/detail", w(handlers.Organization{}.Detail))
		o.GET("/user/getUserInfo", w(handlers.User{}.GetUserByToken))
		o.PUT("/user/updateSelf", w(handlers.User{}.Update))
		o.GET("/systemStatus/search", w(handlers.PortalSystemStatusSearch))
		o.PUT("/consulTags/update", w(handlers.ConsulTagUpdate))
	}

	// IaC管理员权限
	{
		sys := g.Group("/", w(middleware.Auth), w(middleware.IsAdmin))

		//sys.POST("/org/create", w(handlers.Organization{}.Create))
		//sys.PUT("/org/update", w(handlers.Organization{}.Update))
		sys.PUT("/org/changeStatus", w(handlers.Organization{}.ChangeOrgStatus))

		//ctrl.Register(sys.Group("system"), &handlers.SystemConfig{})
		//ctrl.Register(sys.Group("token"), &handlers.Token{})

	}


	{
		root := g.Group("/", w(middleware.Auth), w(middleware.AuthOrgId))
		owner := root.Group("/", w(middleware.IsOrgOwner))
		//owner.GET("/user/search", w(handlers.User{}.Search))
		//owner.GET("/user/detail", w(handlers.User{}.Detail))
		//owner.POST("/user/create", w(handlers.User{}.Create))
		owner.PUT("/user/removeUserForOrg", w(handlers.User{}.RemoveUserForOrg))
		owner.PUT("/user/userPassReset", w(handlers.User{}.UserPassReset))

		//root.PUT("/user/update", w(handlers.User{}.Update))

		root.GET("/gitlab/listRepos", w(handlers.Vcs{}.ListRepos))
		root.GET("/gitlab/listBranches", w(handlers.Vcs{}.ListBranches))
		root.GET("/gitlab/getReadme", w(handlers.Vcs{}.GetReadmeContent))
		//ctrl.Register(root.Group("notification"), &handlers.Notification{})
		ctrl.Register(root.Group("resourceAccount"), &handlers.ResourceAccount{})
		//ctrl.Register(root.Group("template"), &handlers.Template{})

		//ctrl.Register(root.Group("task"), &handlers.Task{})

		ctrl.Register(root.Group("taskComment"), &handlers.TaskComment{})

		//root.GET("/template/overview", w(handlers.Template{}.Overview))
		root.GET("/template/stateSearch", w(handlers.Template{}.Overview))
		//root.GET("/task/last", w(handlers.Task{}.LastTask))

		root.GET("/consulKv/search", w(handlers.ConsulKVSearch))
		root.GET("/runnerList/search", w(handlers.RunnerSearch))
		root.GET("/templateTfvars/search", w(handlers.TemplateTfvarsSearch))
		root.GET("/vcs/listEnableVcs", w(handlers.ListEnableVcs))
		//ctrl.Register(root.Group("vcs"), &handlers.Vcs{})
	}

	//root.GET("/sse/hello/:filename", w(handlers.HelloSse))
	//root.GET("/sse/test", w(handlers.TestSSE))

	//api路径优化v1版本
	o := g.Group("/", w(middleware.Auth))
	{
		o.GET("/org/search", w(handlers.Organization{}.Search))
		o.GET("/org/detail", w(handlers.Organization{}.Detail))
		o.GET("/user/info/search", w(handlers.User{}.GetUserByToken))
		o.PUT("/user/self/update", w(handlers.User{}.Update))
		o.GET("/system/status/search", w(handlers.PortalSystemStatusSearch))
		o.PUT("/consul/tags/update", w(handlers.ConsulTagUpdate))
	}

	// IaC管理员权限
	sys := g.Group("/", w(middleware.Auth), w(middleware.IsAdmin))
	{
		sys.POST("/org/create", w(handlers.Organization{}.Create))
		sys.PUT("/org/update", w(handlers.Organization{}.Update))
		sys.PUT("/org/status/update", w(handlers.Organization{}.ChangeOrgStatus))
		ctrl.Register(sys.Group("system"), &handlers.SystemConfig{})
		ctrl.Register(sys.Group("token"), &handlers.Token{})
	}

	root := g.Group("/", w(middleware.Auth), w(middleware.AuthOrgId))
	owner := root.Group("/", w(middleware.IsOrgOwner))

	{
		owner.GET("/user/search", w(handlers.User{}.Search))
		owner.GET("/user/detail", w(handlers.User{}.Detail))
		owner.POST("/user/create", w(handlers.User{}.Create))
		owner.PUT("/org/user/delete", w(handlers.User{}.RemoveUserForOrg))
		owner.PUT("/user/password/update", w(handlers.User{}.UserPassReset))
	}

	{
		root.PUT("/user/update", w(handlers.User{}.Update))
		ctrl.Register(root.Group("notification"), &handlers.Notification{})
		ctrl.Register(root.Group("resource/account"), &handlers.ResourceAccount{})

		ctrl.Register(root.Group("template"), &handlers.Template{})
		root.GET("/template/overview", w(handlers.Template{}.Overview))
		root.GET("/template/state/search", w(handlers.Template{}.Overview))
		root.GET("/template/tfvars/search", w(handlers.TemplateTfvarsSearch))

		ctrl.Register(root.Group("task"), &handlers.Task{})
		ctrl.Register(root.Group("task/comment"), &handlers.TaskComment{})
		root.GET("/task/last", w(handlers.Task{}.LastTask))

		root.GET("/consul/kv/search", w(handlers.ConsulKVSearch))
		root.GET("/runner/search", w(handlers.RunnerSearch))

		ctrl.Register(root.Group("vcs"), &handlers.Vcs{})
		root.GET("/vcs/repo/search", w(handlers.Vcs{}.ListRepos))
		root.GET("/vcs/branch/search", w(handlers.Vcs{}.ListBranches))
		root.GET("/vcs/readme", w(handlers.Vcs{}.GetReadmeContent))

	}

	// TODO 增加鉴权
	g.GET("/taskLog/sse", w(handlers.Task{}.FollowLogSse))
	g.GET("/task/log/sse", w(handlers.Task{}.FollowLogSse))

}
