package web

import (
	"cloudiac/configs"
	"cloudiac/consts"
	_ "cloudiac/docs" // 千万不要忘了导入你上一步生成的docs
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/utils/logs"
	"cloudiac/web/api"
	api_v1 "cloudiac/web/api/v1"
	"cloudiac/web/api/v1/handlers"
	"cloudiac/web/middleware"
	open_api_v1 "cloudiac/web/openapi/v1"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
)

var logger = logs.Get()

func GetRouter() *gin.Engine {
	w := ctrl.GinRequestCtxWrap

	e := gin.Default()
	// 允许跨域
	e.Use(w(middleware.Cors))
	e.Use(w(middleware.Operation))
	e.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))

	// 普通 handler func
	e.GET("/hello", w(api.Hello))
	e.GET("/system/info", w(func(c *ctx.GinRequestCtx) {
		c.JSONSuccess(gin.H{
			"version": consts.VERSION,
			"build":   consts.BUILD,
		})
	}))
	api_v1.Register(e.Group("/api/v1"))
	open_api_v1.Register(e.Group("/iac/open/v1"))

	e.GET("/template/hook/send", w(handlers.AccessTokenHandler))

	//// http git server
	// 直接提供静态文件访问，生产环境部署时也可以使用 nginx 反代
	e.StaticFS(consts.ReposUrlPrefix, gin.Dir(consts.LocalGitReposPath, true))
	return e
}

func StartServer() {
	conf := configs.Get()
	e := GetRouter()
	logger.Infof("starting server on %v", conf.Listen)
	// API 接口总是使用 http 协议，ssl 证书由 nginx 管理
	if err := e.Run(conf.Listen); err != nil {
		logger.Fatalln(err)
	}
}
