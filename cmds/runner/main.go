package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"cloudiac/cmds/common"
	"cloudiac/configs"
	"cloudiac/runner"
	"cloudiac/utils"
	"cloudiac/utils/consul"
	"cloudiac/utils/logs"

	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
)

type Option struct {
	common.OptionVersion

	Config  string `short:"c" long:"config"  default:"config.yml" description:"config file"`
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug message"`
}

func main() {
	opt := Option{}
	_, err := flags.Parse(&opt)
	if err != nil {
		os.Exit(1)
	}

	common.ShowVersionIf(opt.Version)

	logs.Init(utils.LogLevel(len(opt.Verbose)))
	configs.Init(opt.Config)
	ServiceRegister()
	StartServer()
}

func StartServer() {
	conf := configs.Get()
	logger := logs.Get()

	type request struct {
		*http.Request
		doneCh chan struct{}
	}

	requestChan := make(chan request, 32)
	e := gin.Default()

	apiV1 := e.Group("/api/v1")
	apiV1.Any("/check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
		})
	})
	apiV1.POST("/metrics", func(c *gin.Context) {
		r := request{Request: c.Request, doneCh: make(chan struct{}, 0)}
		requestChan <- r
		<-r.doneCh
	})

	//fp, err := os.OpenFile("./metrics.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	//if err != nil {
	//	panic(err)
	//}

	apiV1.GET("/metrics", func(c *gin.Context) {
		timer := time.NewTimer(time.Millisecond * 100)
		totalRead := int64(0)

		defer func() {
			logger.Debugf("total read %d bytes", totalRead)
			c.Request.Body.Close()
			timer.Stop()
		}()

		for {
			select {
			case <-timer.C:
				return
			case req := <-requestChan:
				//w := io.MultiWriter(c.Writer, fp)
				w := c.Writer
				nr, err := io.Copy(w, req.Body)
				//logger.Infof("copy %d bytes", nr)
				totalRead += nr
				if err != nil {
					logger.Errorf("io copy error: %v", err)
				}
				close(req.doneCh)
			}
		}
	})

	apiV1.POST("/task/run", func(c *gin.Context) {
		logger.Info(c.Request.Body)
		id, err := runner.Run(c.Request)
		if err != nil {
			fmt.Println(err.Error())
			c.JSON(500, gin.H{
				"err": err.Error(),
			})
		} else {
			c.JSON(200, gin.H{
				"id": id,
			})
		}
	})

	apiV1.POST("/task/status", func(c *gin.Context) {
		logger.Debug(c.Request.Body)
		containerStatus, err := runner.Status(c.Request)
		if err != nil {
			logger.Info(err.Error())
			c.JSON(500, gin.H{
				"status": containerStatus.Status.ExitCode,
			})
		} else {
			c.JSON(200, gin.H{
				"status":            containerStatus.Status.Status,
				"status_code":       containerStatus.Status.ExitCode,
				"log_content":       containerStatus.LogContent,
				"log_content_lines": containerStatus.LogContentLines,
			})
		}
	})

	apiV1.POST("/task/cancel", func(c *gin.Context) {
		logger.Debug(c.Request.Body)
		err := runner.Cancel(c.Request)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err,
			})
		} else {
			c.JSON(200, gin.H{
				"error": nil,
			})
		}
	})

	logger.Infof("starting runner on %v", conf.Listen)
	if err := e.Run(conf.Listen); err != nil {
		logger.Fatalln(err)
	}
}

func ServiceRegister() {
	conf := configs.Get()
	logger := logs.Get()

	logger.Debug("Start register CT-Runner service")
	err := consul.Register(conf.Consul)
	if err != nil {
		logger.Debug("Service register failied: %s", err)
	} else {
		logger.Debug("Service register success.")
	}
}
