package web

import (
	"github.com/gin-gonic/gin"

	"cloudiac/configs"
	"cloudiac/consts"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/utils/logs"
	"cloudiac/web/api"
	"cloudiac/web/middleware"

	api_v1 "cloudiac/web/api/v1"
)

var logger = logs.Get()

func StartServer() {
	conf := configs.Get()
	w := ctrl.GinRequestCtxWrap
	e := gin.Default()

	// 允许跨域
	e.Use(w(middleware.Cors))
	//e.Use(w(middleware.Operation))

	// 普通 handler func
	e.GET("/hello", w(api.Hello))
	e.GET("/system/info", w(func(c *ctx.GinRequestCtx) {
		c.JSONSuccess(gin.H{
			"version": consts.VERSION,
			"build":   consts.BUILD,
		})
	}))

	api_v1.Register(e.Group("/api/v1"))

	//// 访问上传静态文件目录
	//e.Static(consts.UploadURLPrefix, conf.UploadDir)
	//// 下载包地址
	//e.Static(consts.DownloadURLPrefix, conf.DownloadDir)

	logger.Infof("starting server on %v", conf.Listen)
	// API 接口总是使用 http 协议，ssl 证书由 nginx 管理
	if err := e.Run(conf.Listen); err != nil {
		logger.Fatalln(err)
	}
}
