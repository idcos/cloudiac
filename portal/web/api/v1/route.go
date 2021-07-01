package v1

import (
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/web/api/v1/handlers"
	"cloudiac/portal/web/middleware"
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

	//api路径优化v1版本
	o := g.Group("/", w(middleware.Auth))
	{
		o.GET("/orgs", w(handlers.Organization{}.Search))
		o.GET("/orgs/:id", w(handlers.Organization{}.Detail))
		o.GET("/user/info/search", w(handlers.User{}.GetUserByToken))
		o.PUT("/user/self/update", w(handlers.User{}.Update))
		o.GET("/system/status/search", w(handlers.PortalSystemStatusSearch))
		o.PUT("/consul/tags/update", w(handlers.ConsulTagUpdate))
		o.GET("/runner/search", w(handlers.RunnerSearch))

	}

	// IaC管理员权限
	sys := g.Group("/", w(middleware.Auth), w(middleware.IsSuperAdmin))
	{
		sys.POST("/orgs", w(handlers.Organization{}.Create))
		sys.PUT("/orgs/:id/status/update", w(handlers.Organization{}.ChangeOrgStatus)) // Deprecated
		sys.PUT("/orgs/:id", w(handlers.Organization{}.Update))
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
		//root.GET("/template/state/search", w(handlers.Template{}.Overview))
		root.GET("/template/tfvars/search", w(handlers.TemplateTfvarsSearch))

		ctrl.Register(root.Group("task"), &handlers.Task{})
		ctrl.Register(root.Group("task/comment"), &handlers.TaskComment{})
		root.GET("/task/last", w(handlers.Task{}.LastTask))

		root.GET("/consul/kv/search", w(handlers.ConsulKVSearch))

		ctrl.Register(root.Group("vcs"), &handlers.Vcs{})
		root.GET("/vcs/repo/search", w(handlers.Vcs{}.ListRepos))
		root.GET("/vcs/branch/search", w(handlers.Vcs{}.ListBranches))
		root.GET("/vcs/readme", w(handlers.Vcs{}.GetReadmeContent))

		ctrl.Register(root.Group("webhook"), &handlers.AccessToken{})
		root.GET("/template/variable/search", w(handlers.TemplateVariableSearch))
		root.GET("/template/playbook/search", w(handlers.TemplatePlaybookSearch))
		root.GET("/template/state_list", w(handlers.Task{}.TaskStateListSearch))
	}

	// TODO 增加鉴权
	g.GET("/task/log/sse", w(handlers.Task{}.FollowLogSse))
}
