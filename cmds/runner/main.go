package main

import (
	"io"
	"net/http"
	"os"
	"time"

	"cloudiac/cmds/common"
	"cloudiac/utils"
	"cloudiac/utils/logs"

	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
)

type Option struct {
	common.OptionVersion

	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug message"`
	Listen  string `short:"l" long:"listen" default:"0.0.0.0:19030" description:"listen address"`
}

func main() {
	opt := Option{}
	_, err := flags.Parse(&opt)
	if err != nil {
		os.Exit(1)
	}
	common.ShowVersionIf(opt.Version)

	logs.Init(utils.LogLevel(len(opt.Verbose)))
	logger := logs.Get()

	type request struct {
		*http.Request
		doneCh chan struct{}
	}

	requestChan := make(chan request, 32)
	e := gin.Default()

	apiV1 := e.Group("/api/v1")
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

	logger.Infof("starting runner on %v", opt.Listen)
	if err := e.Run(opt.Listen); err != nil {
		logger.Fatalln(err)
	}
}
