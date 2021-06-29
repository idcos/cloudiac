package v1

import (
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/web/api/v1/handlers"
	"cloudiac/web/middleware"
	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

// @title cloudIaC 接口文档
// @version 0.3
// @description 统一没有特殊说明的情况下规范说明：<br> 1. 没有特殊说明的情况下，返回数据格式均为`json` <br> 2. 返回 json 最外层均包含`status`和`message`字段 <br> 3. http code [200]: 正常；[400]: 请求参数错误；[401]: 当前请求无token或token过期；[403]: 无访问权限；[404]: 请求资源不存在；[500]: 服务器异常错误；调用方需要对返回http code做好处理。
// @host localhost:9030
// @license.name cloudIaC
// @BasePath /api/v1
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
		o.GET("/org/search", w(handlers.Organization{}.Search))
		o.GET("/org/detail", w(handlers.Organization{}.Detail))
		o.GET("/user/info/search", w(handlers.User{}.GetUserByToken))
		o.PUT("/user/self/update", w(handlers.User{}.Update))
		o.GET("/system/status/search", w(handlers.PortalSystemStatusSearch))
		o.PUT("/consul/tags/update", w(handlers.ConsulTagUpdate))
		o.GET("/runner/search", w(handlers.RunnerSearch))

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
		//root.GET("/template/state/search", w(handlers.Template{}.Overview))
		root.GET("/template/tfvars/search", w(handlers.TemplateTfvarsSearch))

		ctrl.Register(root.Group("task"), &handlers.Task{})
		ctrl.Register(root.Group("task/comment"), &handlers.TaskComment{})
		root.GET("/task/last", w(handlers.Task{}.LastTask))

		root.GET("/consul/kv/search", w(handlers.ConsulKVSearch))

		ctrl.Register(root.Group("vcs"), &handlers.Vcs{})
		//todo 修改api路径与前端联调
		ctrl.Register(root.Group("template/library"), &handlers.MetaTemplate{})
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

	g.GET("/sse/test", w(func(c *ctx.GinRequestCtx) {
		defer c.SSEvent("end", "end")

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		eventId := 0 // to indicate the message id
		for {
			select {
			case <-c.Request.Context().Done():
				return
			case <-ticker.C:
				c.Render(-1, sse.Event{
					Id:    strconv.Itoa(eventId),
					Event: "message",
					Data:  time.Now().Format(time.RFC3339),
				})
				c.Writer.Flush()
				eventId += 1
			}
		}
	}))
}
