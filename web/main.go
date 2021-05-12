package web

import (
	"cloudiac/configs"
	"cloudiac/consts"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/utils/logs"
	"cloudiac/web/api"
	"cloudiac/web/middleware"

	"github.com/gin-gonic/gin"
	"io"
	"os"
	"path/filepath"

	api_v1 "cloudiac/web/api/v1"
	open_api_v1 "cloudiac/web/openapi/v1"
)

var logger = logs.Get()

func GetRouter() *gin.Engine {
	name := "iac-portal"
	abs, _ := filepath.Abs(os.Args[0])
	dir := filepath.Dir(abs)
	ext := filepath.Ext(name)
	execName := name[:len(name)-len(ext)]

	logPath := filepath.Join(dir, "logs", execName+".log")
	f, _ := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND, 0666)
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)


	w := ctrl.GinRequestCtxWrap
	e := gin.Default()

	// 允许跨域
	e.Use(w(middleware.Cors))
	e.Use(w(middleware.Operation))

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

	//// 访问上传静态文件目录
	//e.Static(consts.UploadURLPrefix, conf.UploadDir)
	//// 下载包地址
	//e.Static(consts.DownloadURLPrefix, conf.DownloadDir)
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