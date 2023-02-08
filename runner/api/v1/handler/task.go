// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package handler

import (
	"errors"
	"net/http"

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

	if err := req.Validate(); err != nil {
		c.Error(err, http.StatusBadRequest)
		return
	}

	task := runner.NewTask(req, c.Logger)
	if cid, err := task.Run(); err != nil {
		if errors.Is(err, runner.ErrTaskAborted) {
			c.Result(gin.H{"aborted": true})
		} else {
			c.Error(err, http.StatusInternalServerError)
		}
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

	defer func() {
		_ = runner.CleanTaskWorkDirCode(req.EnvId, req.TaskId)
	}()

	if err := runner.KillContainers(c, req.ContainerIds...); err != nil {
		c.Error(err, http.StatusInternalServerError)
		return
	}
	c.Result(nil)
}
