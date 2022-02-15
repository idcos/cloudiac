// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package task_manager

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/runner"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"net/http"
)

func StopTaskContainers(sess *db.Session, taskId models.Id) error {
	return stopTaskContainers(sess, taskId, false)
}

func StopScanTaskContainers(sess *db.Session, taskId models.Id) error {
	return stopTaskContainers(sess, taskId, true)
}

func stopTaskContainers(sess *db.Session, taskId models.Id, isScanTask bool) error {
	logs.Get().Infof("stop task container, taskId=%s", taskId)

	var (
		runnerId    string
		containerId string
	)
	if isScanTask {
		task, er := services.GetScanTaskById(sess, taskId)
		if er != nil {
			return er
		}
		runnerId = task.RunnerId
		containerId = task.ContainerId
	} else {
		task, er := services.GetTaskById(sess, taskId)
		if er != nil {
			return er
		}
		runnerId = task.RunnerId
		containerId = task.ContainerId
	}

	runnerAddr, err := services.GetRunnerAddress(runnerId)
	if err != nil {
		return err
	}

	requestUrl := utils.JoinURL(runnerAddr, consts.RunnerStopTaskURL)
	req := runner.TaskStopReq{
		TaskId:       taskId.String(),
		ContainerIds: []string{},
	}
	req.ContainerIds = append(req.ContainerIds, containerId)

	header := &http.Header{}
	header.Set("Content-Type", "application/json")
	timeout := int(consts.RunnerConnectTimeout.Seconds())
	_, err = utils.HttpService(requestUrl, "POST", header, req, timeout, timeout)
	return err
}
