// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.
package handler

import (
	"cloudiac/runner"
	"cloudiac/runner/api/ctx"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func AbortTask(c *ctx.Context) {
	req := runner.TaskAbortReq{}
	if err := c.BindJSON(&req); err != nil {
		c.Error(err, http.StatusBadRequest)
		return
	}

	logger := logger.
		WithField("handler", "AbortTask").
		WithField("envId", req.EnvId).
		WithField("taskId", req.TaskId)

	latestStep, err := runner.GetLatestStepInfo(req.EnvId, req.TaskId)
	if err != nil {
		c.Error(err, http.StatusInternalServerError)
		return
	}

	if latestStep.Step < 0 {
		c.Error(fmt.Errorf("cannot stop step '%d'", latestStep.Step), http.StatusInternalServerError)
		return
	}

	logger.Infof("task latest step: %#v", latestStep)
	task, err := runner.LoadStartedTask(req.EnvId, req.TaskId, latestStep.Step)
	if err != nil {
		if os.IsNotExist(err) {
			c.Error(err, http.StatusNotFound)
		} else {
			c.Error(err, http.StatusInternalServerError)
		}
		return
	}

	if info, err := task.ReadControlInfo(); err != nil {
		c.Error(err, http.StatusInternalServerError)
		return
	} else {
		info.AbortedAt = time.Now()
		if err := task.WriteControlInfo(info); err != nil {
			c.Error(err, http.StatusInternalServerError)
			return
		}
	}

	if task.ContainerId != "" {
		// 容器可能被暂停了，所以先启动容器
		if err := (runner.Executor{}).UnpauseIf(task.ContainerId); err != nil {
			c.Error(err, http.StatusInternalServerError)
			return
		}

		if task.ExecId != "" {
			logger.Infof("stop exec: %s", task.ExecId)
			if err := (runner.Executor{}).StopCommand(task.ExecId); err != nil {
				c.Error(err, http.StatusInternalServerError)
				return
			}
		}

		unlockScript := `if cd code/%s && terraform state list; then terraform force-unlock --force %s; else echo 'Not Initialization'; fi`
		if output, err := (runner.Executor{}).RunCommandOutput(task.ContainerId, []string{
			"sh", "-c", fmt.Sprintf(unlockScript, task.Workdir, task.StatePath),
		}); err != nil {
			logger.Errorf("force-unlock error: %v", err)
		} else {
			logger.Infof("forc-unlok output: %s", output)
		}
	}
	c.Result(gin.H{})
}
