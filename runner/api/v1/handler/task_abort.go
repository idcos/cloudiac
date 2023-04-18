// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.
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
	logger.Infof("task latest step: %+v", latestStep)

	info, err := runner.ReadTaskControlInfo(req.EnvId, req.TaskId)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.Warnf("read task control info error: %v", err)
			c.Error(err, http.StatusInternalServerError)
			return
		}
	} else if info.Aborted() {
		c.Error(fmt.Errorf("task aborted"), http.StatusConflict)
		return
	}

	if req.JustCheck { // 只做检查，到这一步就可以返回了
		c.Result(gin.H{"canAbort": true})
		return
	}

	info.EnvId = req.EnvId
	info.TaskId = req.TaskId
	info.AbortedAt = time.Now()
	if err := runner.WriteTaskControlInfo(info); err != nil {
		logger.Warnf("write task control info error: %v", err)
		c.Error(err, http.StatusInternalServerError)
		return
	}

	task, err := runner.LoadStartedTask(req.EnvId, req.TaskId, latestStep.Step)
	if err != nil {
		if os.IsNotExist(err) {
			c.Logger.Debugf("task not started")
			c.Result(gin.H{"aborting": true})
		} else {
			c.Error(err, http.StatusInternalServerError)
		}
		return
	}

	if err := doAbortTask(task); err != nil {
		c.Error(err, http.StatusInternalServerError)
		return
	}
	c.Result(gin.H{"aborting": true})
}

func doAbortTask(task *runner.StartedTask) error {
	if task.ContainerId == "" {
		return nil
	}

	// 容器可能被暂停了，所以先启动容器
	if err := (runner.Executor{}).UnpauseIf(task.ContainerId); err != nil {
		return err
	}

	if task.ExecId != "" {
		logger.Infof("stop exec: %s", task.ExecId)
		if err := (runner.Executor{}).StopCommand(task.ExecId); err != nil {
			return err
		}
	}

	unlockScript := `if cd code/%s && terraform state list; then terraform force-unlock --force %s; else echo 'Not Initialization'; fi`
	if output, err := (runner.Executor{}).RunCommandOutput(task.ContainerId, []string{
		"sh", "-c", fmt.Sprintf(unlockScript, task.Workdir, task.StatePath),
	}); err != nil {
		logger.Errorf("force-unlock error: %v", err)
	} else {
		logger.Infof("forc-unlock output: %s", output)
	}
	return nil
}
