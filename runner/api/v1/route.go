package v1

import (
	"cloudiac/runner/api/ctx"
	"cloudiac/runner/api/v1/handler"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"cloudiac/utils/logs"
)

func RegisterRoute(apiV1 *gin.RouterGroup) {
	w := ctx.HandlerWrapper

	logger := logs.Get()
	apiV1.Any("/check", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
		})
	})

	type request struct {
		*http.Request
		doneCh chan struct{}
	}

	requestChan := make(chan request, 32)
	apiV1.POST("/metrics", func(c *gin.Context) {
		r := request{Request: c.Request, doneCh: make(chan struct{}, 0)}
		requestChan <- r
		<-r.doneCh
	})

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

	apiV1.POST("/task/run", w(handler.RunTask))
	apiV1.GET("/task/status", w(handler.TaskStatus))
	apiV1.GET("/task/log/follow", w(handler.TaskLogFollow))

	//apiV1.POST("/task/cancel", func(c *gin.Context) {
	//	logger.Debug(c.Request.Body)
	//	err := runner.Cancel(c.Request)
	//	if err != nil {
	//		c.JSON(500, gin.H{
	//			"error": err,
	//		})
	//	} else {
	//		c.JSON(200, gin.H{
	//			"error": nil,
	//		})
	//	}
	//})
}
