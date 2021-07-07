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
	a := middleware.AccessControl

	// 非授权用户相关路由
	g.Any("/check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
		})
	})
	g.POST("/auth/login", w(handlers.Auth{}.Login))

	// TODO 增加鉴权
	g.GET("/task/log/sse", w(handlers.Task{}.FollowLogSse))

	// Authorization Header 鉴权
	g.Use(w(middleware.Auth)) // 解析 header token

	ctrl.Register(g.Group("token", a()), &handlers.Auth{})
	ctrl.Register(g.Group("system", a()), &handlers.SystemConfig{})
	ctrl.Register(g.Group("webhook", a()), &handlers.AccessToken{})
	g.GET("/auth/me", a("self", "read"), w(handlers.Auth{}.GetUserByToken))
	g.PUT("/users/self", a("self", "update"), w(handlers.User{}.UpdateSelf))
	g.GET("/runner/search", a(), w(handlers.RunnerSearch))
	g.PUT("/consul/tags/update", a(), w(handlers.ConsulTagUpdate))
	g.GET("/consul/kv/search", a(), w(handlers.ConsulKVSearch))
	g.GET("/system/status/search", a(), w(handlers.PortalSystemStatusSearch))

	ctrl.Register(g.Group("orgs", a()), &handlers.Organization{})
	g.PUT("/orgs/:id/status", a(), w(handlers.Organization{}.ChangeOrgStatus))
	g.GET("/orgs/:id/users", a("orgs", "listuser"), w(handlers.Organization{}.SearchUser))
	g.PUT("/orgs/:id/users", a("orgs", "adduser"), w(handlers.Organization{}.AddUserToOrg))
	g.DELETE("/orgs/:id/users/:userId", a("orgs", "removeuser"), w(handlers.Organization{}.RemoveUserForOrg))
	g.PUT("/orgs/:id/users/:userId/role", a("orgs", "updaterole"), w(handlers.Organization{}.UpdateUserOrgRel))

	// 组织 header
	g.Use(w(middleware.AuthOrgId))

	ctrl.Register(g.Group("users", a()), &handlers.User{})
	g.PUT("/users/:id/status", a(), w(handlers.User{}.ChangeUserStatus))
	g.POST("/users/:id/password/reset", a(), w(handlers.User{}.PasswordReset))

	g.POST("/projects", a(), w(handlers.Organization{}.Search))
	g.GET("/projects", a(), w(handlers.Organization{}.Search))
	g.GET("/projects/:id", a(), w(handlers.Organization{}.Search))

	// 项目资源
	// TODO: parse project header
	g.Use(w(middleware.AuthProjectId))

	ctrl.Register(g.Group("template", a()), &handlers.Template{})
	g.GET("/template/overview", a(), w(handlers.Template{}.Overview))
	g.GET("/template/tfvars/search, a()", w(handlers.TemplateTfvarsSearch))
	g.GET("/template/variable/search", a(), w(handlers.TemplateVariableSearch))
	g.GET("/template/playbook/search", a(), w(handlers.TemplatePlaybookSearch))
	g.GET("/template/state_list", a(), w(handlers.Task{}.TaskStateListSearch))

	ctrl.Register(g.Group("task", a()), &handlers.Task{})
	ctrl.Register(g.Group("task/comment", a()), &handlers.TaskComment{})
	g.GET("/task/last", a(), w(handlers.Task{}.LastTask))

	ctrl.Register(g.Group("vcs", a()), &handlers.Vcs{})
	g.GET("/vcs/repo/search", a(), w(handlers.Vcs{}.ListRepos))
	g.GET("/vcs/branch/search", a(), w(handlers.Vcs{}.ListBranches))
	g.GET("/vcs/readme", a(), w(handlers.Vcs{}.GetReadmeContent))

	ctrl.Register(g.Group("notification", a()), &handlers.Notification{})
	ctrl.Register(g.Group("resource/account", a()), &handlers.ResourceAccount{})
}
