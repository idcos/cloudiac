// Copyright 2021 CloudJ Company Limited. All rights reserved.

package v1

import (
	"cloudiac/common"
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

// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey AuthToken
// @in header
// @name Authorization

func Register(g *gin.RouterGroup) {
	w := ctrl.WrapHandler
	ac := middleware.AccessControl

	// 非授权用户相关路由
	g.Any("/check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
			"version": common.VERSION,
			"build":   common.BUILD,
		})
	})

	g.Use(gin.Logger())

	g.POST("/trigger/send", w(handlers.ApiTriggerHandler))

	// sso token 验证
	g.GET("/sso/tokens/verify", w(handlers.VerifySsoToken))

	// 触发器
	apiToken := g.Group("")
	apiToken.Use(w(middleware.AuthApiToken))
	apiToken.POST("/webhooks/:vcsType/:vcsId", w(handlers.WebhooksApiHandler))

	g.POST("/auth/login", w(handlers.Auth{}.Login))

	// Authorization Header 鉴权
	g.Use(w(middleware.Auth)) // 解析 header token

	// 创建单点登录 token
	g.POST("/sso/tokens", w(handlers.GenerateSsoToken))
	// ctrl.Register(g.Group("tokens", ac()), &handlers.Token{})

	// TODO 旧的 token 接口己不再使用，先注释, 后续版本删除
	// ctrl.Register(g.Group("token", ac()), &handlers.Auth{})

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
	g.PUT("/systems", ac(), w(handlers.SystemConfig{}.Update))
	g.GET("/systems", ac(), w(handlers.SystemConfig{}.Search))
	// 系统状态
	g.GET("/systems/status", w(handlers.PortalSystemStatusSearch))
	// 系统设置registry addr 配置
	g.GET("/system_config/registry/addr", ac(), w(handlers.GetRegistryAddr))     // 获取registry地址的设置
	g.POST("/system_config/registry/addr", ac(), w(handlers.UpsertRegistryAddr)) // 更新registry地址的设置

	// 要求组织 header
	g.Use(w(middleware.AuthOrgId))

	// 策略管理
	ctrl.Register(g.Group("policies", ac()), &handlers.Policy{})
	g.GET("/policies/summary", ac(), w(handlers.Policy{}.PolicySummary))
	g.GET("/policies/:id/error", ac(), w(handlers.Policy{}.PolicyError))
	g.GET("/policies/:id/suppress", ac(), w(handlers.Policy{}.SearchPolicySuppress))
	g.POST("/policies/:id/suppress", ac("suppress"), w(handlers.Policy{}.UpdatePolicySuppress))
	g.GET("/policies/:id/suppress/sources", ac(), w(handlers.Policy{}.SearchPolicySuppressSource))
	g.DELETE("/policies/:id/suppress/:suppressId", ac("suppress"), w(handlers.Policy{}.DeletePolicySuppress))
	g.GET("/policies/:id/report", ac(), w(handlers.Policy{}.PolicyReport))
	g.POST("/policies/parse", ac(), w(handlers.Policy{}.Parse))
	g.POST("/policies/test", ac(), w(handlers.Policy{}.Test))
	g.GET("/policies/templates", ac(), w(handlers.Policy{}.SearchPolicyTpl))
	g.PUT("/policies/templates/:id", ac(), w(handlers.Policy{}.UpdatePolicyTpl))
	g.PUT("/policies/templates/:id/enabled", ac("enablescan"), w(handlers.Policy{}.EnablePolicyTpl))
	g.GET("/policies/templates/:id/policies", ac(), w(handlers.Policy{}.TplOfPolicy))
	g.GET("/policies/templates/:id/groups", ac(), w(handlers.Policy{}.TplOfPolicyGroup))
	g.GET("/policies/templates/:id/valid_policies", ac(), w(handlers.Policy{}.ValidTplOfPolicy))
	g.POST("/policies/templates/:id/scan", ac("scan"), w(handlers.Policy{}.ScanTemplate))
	g.POST("/policies/templates/scans", ac("scan"), w(handlers.Policy{}.ScanTemplates))
	g.GET("/policies/templates/:id/result", ac(), w(handlers.Policy{}.TemplateScanResult))
	g.GET("/policies/envs", ac(), w(handlers.Policy{}.SearchPolicyEnv))
	g.PUT("/policies/envs/:id", ac(), w(handlers.Policy{}.UpdatePolicyEnv))
	g.PUT("/policies/envs/:id/enabled", ac("enablescan"), w(handlers.Policy{}.EnablePolicyEnv))
	g.GET("/policies/envs/:id/policies", ac(), w(handlers.Policy{}.EnvOfPolicy))
	g.GET("/policies/envs/:id/valid_policies", ac(), w(handlers.Policy{}.ValidEnvOfPolicy))
	g.POST("/policies/envs/:id/scan", ac("scan"), w(handlers.Policy{}.ScanEnvironment))
	g.GET("/policies/envs/:id/result", ac(), w(handlers.Policy{}.EnvScanResult))
	ctrl.Register(g.Group("policies/groups", ac()), &handlers.PolicyGroup{})
	g.POST("/policies/groups/checks", ac(), w(handlers.PolicyGroupChecks))
	g.GET("/policies/groups/:id/policies", ac(), w(handlers.PolicyGroup{}.SearchGroupOfPolicy))
	g.POST("/policies/groups/:id", ac(), w(handlers.PolicyGroup{}.OpPolicyAndPolicyGroupRel))
	g.GET("/policies/groups/:id/report", ac(), w(handlers.PolicyGroup{}.ScanReport))
	g.GET("/policies/groups/:id/last_tasks", ac(), w(handlers.PolicyGroup{}.LastTasks))

	// Registry 侧策略组API
	g.GET("/registry/policy_groups", ac(), w(handlers.SearchRegistryPolicyGroups))
	g.GET("/registry/policy_group/versions", ac(), w(handlers.SearchRegistryPolicyGroupVersions))

	// 组织下的资源搜索(只需要有环境的读权限即可查看资源)
	g.GET("/orgs/resources", ac("envs", "read"), w(handlers.Organization{}.SearchOrgResources))

	// 组织用户管理
	g.GET("/orgs/:id/users", ac("orgs", "listuser"), w(handlers.Organization{}.SearchUser))
	g.POST("/orgs/:id/users", ac("orgs", "adduser"), w(handlers.Organization{}.AddUserToOrg))
	g.PUT("/orgs/:id/users/:userId/role", ac("orgs", "updaterole"), w(handlers.Organization{}.UpdateUserOrgRel))
	g.PUT("/orgs/:id/users/:userId", ac("orgs", "updaterole"), w(handlers.Organization{}.UpdateUserOrg))
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
	g.PUT("/variables/scope/:scope/:id", ac(), w(handlers.Variable{}.UpdateObjectVars))
	// 供第三方系统获取变量的接口，该接口将 terraform 变量和环境变量统一转为环境变量格式返回，方便第三方系统处理
	g.GET("/variables/sample", ac(), w(handlers.Variable{}.SearchSampleVariable))
	ctrl.Register(g.Group("variables", ac()), &handlers.Variable{})

	// 变量组
	ctrl.Register(g.Group("var_groups", ac()), &handlers.VariableGroup{})
	g.GET("/var_groups/relationship", ac(), w(handlers.VariableGroup{}.SearchRelationship))
	g.GET("/var_groups/relationship/all", ac(), w(handlers.VariableGroup{}.SearchRelationshipAll))
	g.PUT("/var_groups/relationship/batch", ac(), w(handlers.VariableGroup{}.BatchUpdateRelationship))
	//g.DELETE("/var_groups/relationship/:id", ac(), w(handlers.VariableGroup{}.DeleteRelationship))

	//token管理
	ctrl.Register(g.Group("tokens", ac()), &handlers.Token{})
	//密钥管理
	ctrl.Register(g.Group("keys", ac()), &handlers.Key{})

	ctrl.Register(g.Group("vcs", ac()), &handlers.Vcs{})
	g.GET("/vcs/:id/repo", ac(), w(handlers.Vcs{}.ListRepos))
	g.GET("/vcs/:id/branch", ac(), w(handlers.Vcs{}.ListBranches))
	g.GET("/vcs/:id/tag", ac(), w(handlers.Vcs{}.ListTags))
	g.GET("/vcs/:id/readme", ac(), w(handlers.Vcs{}.GetReadmeContent))
	// 云模板
	ctrl.Register(g.Group("templates", ac()), &handlers.Template{})
	g.GET("/templates/variables", ac(), w(handlers.TemplateVariableSearch))
	g.GET("/templates/tfversions", ac(), w(handlers.TemplateTfVersionSearch))
	g.GET("/templates/autotfversion", ac(), w(handlers.AutoTemplateTfVersionChoice))
	g.POST("/templates/checks", ac(), w(handlers.TemplateChecks))
	g.GET("/templates/export", ac(), w(handlers.TemplateExport))
	g.POST("/templates/import", ac(), w(handlers.TemplateImport))
	g.GET("/vcs/:id/repos/tfvars", ac(), w(handlers.TemplateTfvarsSearch))
	g.GET("/vcs/:id/repos/playbook", ac(), w(handlers.TemplatePlaybookSearch))
	g.GET("/vcs/:id/file", ac(), w(handlers.Vcs{}.SearchVcsFileContent))
	ctrl.Register(g.Group("notifications", ac()), &handlers.Notification{})

	// 任务实时日志（云模板检测无项目ID）
	g.GET("/tasks/:id/log/sse", ac(), w(handlers.Task{}.FollowLogSse))

	// 项目资源
	g.Use(w(middleware.AuthProjectId))

	// 环境管理
	ctrl.Register(g.Group("envs", ac()), &handlers.Env{})
	g.PUT("/envs/:id/archive", ac(), w(handlers.Env{}.Archive))
	g.GET("/envs/:id/tasks", ac(), w(handlers.Env{}.SearchTasks))
	g.GET("/envs/:id/tasks/last", ac(), w(handlers.Env{}.LastTask))
	g.POST("/envs/:id/deploy", ac("envs", "deploy"), w(handlers.Env{}.Deploy))
	g.POST("/envs/:id/destroy", ac("envs", "destroy"), w(handlers.Env{}.Destroy))
	g.GET("/envs/:id/resources", ac(), w(handlers.Env{}.SearchResources))
	g.GET("/envs/:id/output", ac(), w(handlers.Env{}.Output))
	g.GET("/envs/:id/resources/:resourceId", ac(), w(handlers.Env{}.ResourceDetail))
	g.GET("/envs/:id/variables", ac(), w(handlers.Env{}.Variables))
	g.GET("/envs/:id/policy_result", ac(), w(handlers.Env{}.PolicyResult))
	g.GET("/envs/:id/resources/graph", ac(), w(handlers.Env{}.SearchResourcesGraph))
	g.GET("/envs/:id/resources/graph/:resourceId", ac(), w(handlers.Env{}.ResourceGraphDetail))

	// 任务管理
	g.GET("/tasks", ac(), w(handlers.Task{}.Search))
	g.GET("/tasks/:id", ac(), w(handlers.Task{}.Detail))
	g.GET("/tasks/:id/log", ac(), w(handlers.Task{}.Log))
	g.GET("/tasks/:id/output", ac(), w(handlers.Task{}.Output))
	g.GET("/tasks/:id/resources", ac(), w(handlers.Task{}.Resource))
	g.POST("/tasks/:id/approve", ac("tasks", "approve"), w(handlers.Task{}.TaskApprove))
	g.POST("/tasks/:id/comment", ac(), w(handlers.TaskComment{}.Create))
	g.GET("/tasks/:id/comment", ac(), w(handlers.TaskComment{}.Search))
	g.GET("/tasks/:id/steps", ac(), w(handlers.Task{}.SearchTaskStep))
	g.GET("/tasks/:id/steps/:stepId/log", ac(), w(handlers.Task{}.GetTaskStepLog))
	g.GET("/tasks/:id/steps/:stepId/log/sse", ac(), w(handlers.Task{}.FollowStepLogSse))
	g.GET("/tasks/:id/resources/graph", ac(), w(handlers.Task{}.ResourceGraph))

	//g.GET("/tokens/trigger", ac(), w(handlers.Token{}.VcsWebhookUrl))
	g.GET("/vcs/webhook", ac(), w(handlers.Token{}.VcsWebhookUrl))
	ctrl.Register(g.Group("resource/account", ac()), &handlers.ResourceAccount{})
}
