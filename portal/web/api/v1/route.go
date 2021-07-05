package v1

import (
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/web/api/v1/handlers"
	"cloudiac/portal/web/middleware"
	"github.com/gin-gonic/gin"
)

// @title 云霁 CloudIaC 基础设施即代码管理平台
// @version 1.0.0
// @description CloudIaC 是基于基础设施即代码构建的云环境自动化管理平台。CloudIaC 将易于使用的界面与强大的治理工具相结合，让您和您团队的成员可以快速轻松的在云中部署和管理环境。 <br />通过将 CloudIaC 集成到您的流程中，您可以获得对组织的云使用情况的可见性、可预测性和更好的治理。

// @host localhost:9030
// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey AuthToken
// @in header
// @name Authorization

func Register(g *gin.RouterGroup) {
	w := ctrl.GinRequestCtxWrap

	// 非授权用户相关路由
	g.Any("/check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
		})
	})
	auth := g.Group("/auth")
	auth.POST("/login", w(handlers.Auth{}.Login))

	// 所有登陆用户相关路由
	o := g.Group("/", w(middleware.Auth))
	{
		o.GET("/orgs", w(handlers.Organization{}.Search))
		o.GET("/orgs/:id", w(handlers.Organization{}.Detail))

		o.PUT("/users/self", w(handlers.User{}.UpdateSelf))

		o.GET("/token/me", w(handlers.Auth{}.GetUserByToken))
		o.GET("/system/status/search", w(handlers.PortalSystemStatusSearch))
		o.PUT("/consul/tags/update", w(handlers.ConsulTagUpdate))
		o.GET("/runner/search", w(handlers.RunnerSearch))
	}

	// 平台管理员权限
	sys := g.Group("/", w(middleware.Auth), w(middleware.IsSuperAdmin))
	{
		sys.POST("/orgs", w(handlers.Organization{}.Create))
		sys.PUT("/orgs/:id", w(handlers.Organization{}.Update))
		sys.PUT("/orgs/:id/status", w(handlers.Organization{}.ChangeOrgStatus))
		sys.DELETE("/orgs/:id", w(handlers.Organization{}.Delete))

		auth.PUT("/auth/password", w(handlers.Auth{}.UserPassReset))

		ctrl.Register(sys.Group("system"), &handlers.SystemConfig{})
		ctrl.Register(sys.Group("token"), &handlers.Auth{})
	}

	// 组织成员相关路由
	root := g.Group("/", w(middleware.Auth), w(middleware.AuthOrgId))

	// 组织管理员相关路由
	orgAdmin := root.Group("/", w(middleware.IsOrgAdmin))
	{
		orgAdmin.POST("/users", w(handlers.User{}.Create))
		orgAdmin.PUT("/users/:id", w(handlers.User{}.Update))
		orgAdmin.PUT("/users/:id/status", w(handlers.User{}.ChangeUserStatus))
		orgAdmin.DELETE("/users/:id", w(handlers.User{}.Delete))

		orgAdmin.DELETE("/orgs/:id/users/:userId", w(handlers.Organization{}.RemoveUserForOrg))
		orgAdmin.PUT("/orgs/:id/users", w(handlers.Organization{}.AddUserToOrg))
		orgAdmin.PUT("/orgs/:id/users/:userId/role", w(handlers.Organization{}.UpdateUserOrgRel))
	}

	// 组织普通用户相关路由
	{
		root.GET("/users", w(handlers.User{}.Search))
		root.GET("/users/:id", w(handlers.User{}.Detail))

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

	//项目管理
	{
		root.GET("/project", w(handlers.Project{}.Search))
		root.POST("/project", w(handlers.Project{}.Create))
		root.PUT("/project/:id", w(handlers.Project{}.Update))
		root.DELETE("/project/:id", w(handlers.Project{}.Delete))
		root.GET("/project/:id", w(handlers.Project{}.Detail))
	}

	// TODO 增加鉴权
	g.GET("/task/log/sse", w(handlers.Task{}.FollowLogSse))
}
