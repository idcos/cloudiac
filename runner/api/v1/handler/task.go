// Copyright 2021 CloudJ Company Limited. All rights reserved.

package handler

import (
	"net/http"

	"github.com/docker/docker/api/types"
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

	containerRemoveOpts := types.ContainerRemoveOptions{
		Force: true,
	}

	for _, cid := range req.ContainerIds {
		if err := cli.ContainerRemove(c.Context, cid, containerRemoveOpts); err != nil {
			if _, ok := err.(errdefs.ErrNotFound); ok {
				continue
			}
			c.Error(err, http.StatusInternalServerError)
			return
		}
	}
}
