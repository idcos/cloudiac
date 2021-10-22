package task_manager

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/runner"
	"cloudiac/utils"
	"net/http"
)

func StopTaskContainers(sess *db.Session, taskId models.Id) error {
	task, er := services.GetTaskById(sess, taskId)
	if er != nil {
		return er
	}

	taskJobs, er := services.GetTaskJobs(sess, taskId)
	if er != nil {
		return er
	}

	runnerAddr, err := services.GetRunnerAddress(task.RunnerId)
	if err != nil {
		return err
	}

	requestUrl := utils.JoinURL(runnerAddr, consts.RunnerStopTaskURL)
	req := runner.TaskStopReq{
		TaskId:       taskId.String(),
		ContainerIds: []string{},
	}
	for _, job := range taskJobs {
		req.ContainerIds = append(req.ContainerIds, job.ContainerId)
	}

	header := &http.Header{}
	header.Set("Content-Type", "application/json")
	timeout := int(consts.RunnerConnectTimeout.Seconds())
	_, err = utils.HttpService(requestUrl, "POST", header, req, timeout, timeout)
	return err
}
