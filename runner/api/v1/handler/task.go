// Copyright 2021 CloudJ Company Limited. All rights reserved.

package handler

import (
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

	task := runner.NewTask(req, c.Logger)
	if cid, err := task.Run(); err != nil {
		c.Error(err, http.StatusInternalServerError)
		return
	} else {
		c.Result(gin.H{"containerId": cid})
	}
}
