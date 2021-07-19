package v1

import (
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/web/api/v1/handlers"
	"cloudiac/portal/web/middleware"
	"github.com/gin-gonic/gin"
)

// @title 云霁 CloudIaC 基础设施即代码管理平台
// @version 1.0.0
// @description CloudIaC 是基于基础设施即代码构建的云环境自动化管理平台。
// @description CloudIaC 将易于使用的界面与强大的治理工具相结合，让您和您团队的成员可以快速轻松的在云中部署和管理环境。
// @description 通过将 CloudIaC 集成到您的流程中，您可以获得对组织的云使用情况的可见性、可预测性和更好的治理。

// @host localhost:9030
// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey AuthToken
// @in header
// @name Authorization

func Register(g *gin.RouterGroup) {
	w := ctrl.GinRequestCtxWrap
	ac := middleware.AccessControl

	// 非授权用户相关路由
	g.Any("/check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
		})
	})
	g.GET("/trigger/send", w(handlers.ApiTriggerHandler))

	g.POST("/auth/login", w(handlers.Auth{}.Login))

	// TODO 增加鉴权
	g.GET("/task/log/sse", w(handlers.Task{}.FollowLogSse))

	// Authorization Header 鉴权
	g.Use(w(middleware.Auth)) // 解析 header token

	ctrl.Register(g.Group("token", ac()), &handlers.Auth{})
	g.GET("/auth/me", ac("self", "read"), w(handlers.Auth{}.GetUserByToken))
	g.PUT("/users/self", ac("self", "update"), w(handlers.User{}.UpdateSelf))
	//todo runner list权限怎么划分
	g.GET("/runners", ac(), w(handlers.RunnerSearch))
	g.PUT("/consul/tags/update", ac(), w(handlers.ConsulTagUpdate))
	g.GET("/consul/kv/search", ac(), w(handlers.ConsulKVSearch))

	ctrl.Register(g.Group("orgs", ac()), &handlers.Organization{})
	g.PUT("/orgs/:id/status", ac(), w(handlers.Organization{}.ChangeOrgStatus))
	ctrl.Register(g.Group("users", ac()), &handlers.User{})
	g.PUT("/users/:id/status", ac(), w(handlers.User{}.ChangeUserStatus))
	g.POST("/users/:id/password/reset", ac(), w(handlers.User{}.PasswordReset))

	// 系统配置
	ctrl.Register(g.Group("systems"), &handlers.SystemConfig{})
	// 系统状态
	g.GET("/systems/status", w(handlers.PortalSystemStatusSearch))

	// 要求组织 header
	g.Use(w(middleware.AuthOrgId))

	// 组织用户管理
	g.GET("/orgs/:id/users", ac("orgs", "listuser"), w(handlers.Organization{}.SearchUser))
	g.POST("/orgs/:id/users", ac("orgs", "adduser"), w(handlers.Organization{}.AddUserToOrg))
	g.PUT("/orgs/:id/users/:userId/role", ac("orgs", "updaterole"), w(handlers.Organization{}.UpdateUserOrgRel))
	g.POST("/orgs/:id/users/invite", ac("orgs", "adduser"), w(handlers.Organization{}.InviteUser))
	g.DELETE("/orgs/:id/users/:userId", ac("orgs", "removeuser"), w(handlers.Organization{}.RemoveUserForOrg))

	g.GET("/projects/users", ac(), w(handlers.ProjectUser{}.Search))
	g.GET("/projects/authorization/users", ac(), w(handlers.ProjectUser{}.SearchProjectAuthorizationUser))
	g.POST("/projects/users", ac(), w(handlers.ProjectUser{}.Create))
	g.PUT("/projects/users/:id", ac(), w(handlers.ProjectUser{}.Update))
	g.DELETE("/projects/users/:id", ac(), w(handlers.ProjectUser{}.Delete))

	//项目管理
	ctrl.Register(g.Group("projects", ac()), &handlers.Project{})
	//变量管理
	g.PUT("/variables/batch", ac(), w(handlers.Variable{}.BatchUpdate))
	ctrl.Register(g.Group("variables", ac()), &handlers.Variable{})
	//token管理
	ctrl.Register(g.Group("tokens", ac()), &handlers.Token{})
	//密钥管理
	ctrl.Register(g.Group("keys", ac()), &handlers.Key{})

	ctrl.Register(g.Group("vcs", ac()), &handlers.Vcs{})
	g.GET("/vcs/:id/repo", ac(), w(handlers.Vcs{}.ListRepos))
	g.GET("/vcs/:id/branch", ac(), w(handlers.Vcs{}.ListBranches))
	g.GET("/vcs/:id/tag", ac(), w(handlers.Vcs{}.ListTags))
	g.GET("/vcs/:id/readme", ac(), w(handlers.Vcs{}.GetReadmeContent))
	ctrl.Register(g.Group("templates", ac()), &handlers.Template{})
	g.GET("/templates/variables", ac(), w(handlers.TemplateVariableSearch))
	g.GET("/vcs/:id/repos/tfvars", ac(), w(handlers.TemplateTfvarsSearch))
	g.GET("/vcs/:id/repos/playbook", ac(), w(handlers.TemplatePlaybookSearch))
	ctrl.Register(g.Group("notifications", ac()), &handlers.Notification{})

	// 项目资源
	// TODO: parse project header
	g.Use(w(middleware.AuthProjectId))

	ctrl.Register(g.Group("envs", ac()), &handlers.Env{})
	g.PUT("/envs/:id/archive", ac(), w(handlers.Env{}.Archive))
	g.GET("/envs/:id/tasks", ac(), w(handlers.Env{}.SearchTasks))
	g.GET("/envs/:id/tasks/last", ac(), w(handlers.Env{}.LastTask))
	g.POST("/envs/:id/deploy", ac("envs", "deploy"), w(handlers.Env{}.Deploy))
	g.POST("/envs/:id/destroy", ac("envs", "destroy"), w(handlers.Env{}.Destroy))
	g.GET("/envs/:id/resources", ac(), w(handlers.Env{}.SearchResources))
	g.GET("/envs/:id/variables", ac(), w(handlers.Env{}.SearchVariables))

	g.GET("/tasks", ac(), w(handlers.Task{}.Search))
	g.GET("/tasks/:id", ac(), w(handlers.Task{}.Detail))
	g.GET("/tasks/:id/log", ac(), w(handlers.Task{}.Log))
	g.GET("/tasks/:id/output", ac(), w(handlers.Task{}.Output))
	g.POST("/tasks/:id/approve", ac("tasks", "approve"), w(handlers.Task{}.TaskApprove))
	g.POST("/tasks/:id/comment", ac(), w(handlers.TaskComment{}.Create))
	g.GET("/tasks/:id/comment", ac(), w(handlers.TaskComment{}.Search))

	g.GET("/tokens/trigger", ac(), w(handlers.Token{}.DetailTriggerToken))
	ctrl.Register(g.Group("resource/account", ac()), &handlers.ResourceAccount{})
}
