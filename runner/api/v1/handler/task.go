// Copyright 2021 CloudJ Company Limited. All rights reserved.

package handler

import (
	"net/http"
	"strings"

	"github.com/docker/docker/errdefs"
	"github.com/gin-gonic/gin"

	"cloudiac/runner"
	"cloudiac/runner/api/ctx"
)

func RunTask(c *ctx.Context) {
	req := runner.RunTaskReq{}
	if err := c.BindJSON(&req); err != nil {
		c.Error(err, http.StatusBadRequest)
		return
	}

	task := runner.NewTask(req, c.Logger)
	if cid, err := task.Run(); err != nil {
		c.Error(err, http.StatusInternalServerError)
		return
	} else {
		c.Result(gin.H{"containerId": cid})
	}
}

func StopTask(c *ctx.Context) {
	req := runner.TaskStopReq{}
	if err := c.BindJSON(&req); err != nil {
		c.Error(err, http.StatusBadRequest)
		return
	}

	cli, err := runner.DockerClient()
	if err != nil {
		c.Error(err, http.StatusInternalServerError)
		return
	}

	// 这里仅 kill container，container 的 remove 通过启动时的 AutoRemove 参数配置
	for _, cid := range req.ContainerIds {
		// default signal "SIGKILL"
		if err := cli.ContainerKill(c.Context, cid, ""); err != nil {
			if _, ok := err.(errdefs.ErrNotFound); ok {
				continue
			}

			if err != nil {
				// 有可能己经提交了删除请求，这里忽略掉这些报错
				if !strings.Contains(err.Error(), "already in progress") &&
					!strings.Contains(err.Error(), "No such container") {
					logger.Warnf("kill container error: %v", err)
				}
			}

			c.Error(err, http.StatusInternalServerError)
			return
		}
	}
	c.Result(nil)
}
