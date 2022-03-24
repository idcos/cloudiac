// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package task_manager

import (
	"cloudiac/portal/libs/db"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"cloudiac/portal/consts"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/portal/services/logstorage"
	"cloudiac/runner"
	"cloudiac/utils"
	"cloudiac/utils/logs"
)

// StartTaskStep 启动任务的一步
// 该函数会设置 taskReq 中 step 相关的数据
func StartTaskStep(taskReq runner.RunTaskReq, step models.TaskStep) (
	containerId string, retryAble bool, err error) {

	logger := logs.Get().
		WithField("action", "StartTaskStep").
		WithField("taskId", taskReq.TaskId).
		WithField("step", step.Index)

	header := &http.Header{}
	header.Set("Content-Type", "application/json")

	var runnerAddr string
	runnerAddr, err = services.GetRunnerAddress(taskReq.RunnerId)
	if err != nil {
		return "", true, err
	}

	requestUrl := utils.JoinURL(runnerAddr, consts.RunnerRunTaskStepURL)
	logger.Debugf("request runner: %s", requestUrl)

	taskReq.Step = step.Index
	taskReq.StepType = step.Type
	taskReq.StepArgs = step.Args

	respData, err := utils.HttpService(requestUrl, "POST", header, taskReq,
		int(consts.RunnerConnectTimeout.Seconds()), int(consts.RunnerConnectTimeout.Seconds())*10)
	if err != nil {
		return "", true, err
	}

	resp := runner.Response{}
	if err := json.Unmarshal(respData, &resp); err != nil {
		return "", false, fmt.Errorf("unexpected response: %s", respData)
	}
	logger.Debugf("runner response: %s", respData)

	if resp.Error != "" {
		return "", false, fmt.Errorf(resp.Error)
	}

	if result, ok := resp.Result.(map[string]interface{}); !ok {
		return "", false, fmt.Errorf("unexpected result: %v", resp.Result)
	} else {
		containerId = fmt.Sprintf("%v", result["containerId"])
	}

	return containerId, false, nil
}

type waitStepResult struct {
	Status string
	Result runner.TaskStatusMessage
}

// WaitTaskStep 等待任务结束(包括超时)，返回任务最新状态
// 该函数会更新任务状态、日志等到 db
func WaitTaskStep(ctx context.Context, sess *db.Session, task *models.Task, step *models.TaskStep) (
	stepResult *waitStepResult, err error) {
	logger := logs.Get().WithField("action", "WaitTaskStep").WithField("taskId", task.Id)
	if step.StartAt == nil {
		return nil, fmt.Errorf("step not start")
	}

	// runner 端己经增加了超时处理，portal 端的超时暂时保留，但时间设置为给定时间的 2 倍
	taskDeadline := time.Time(*step.StartAt).Add(time.Duration(task.StepTimeout*2) * time.Second)

	// 当前版本实现中需要 portal 主动连接到 runner 获取状态
	err = utils.RetryFunc(10, time.Second*5, func(retryN int) (retry bool, er error) {
		stepResult, er = pullTaskStepStatus(ctx, task, step, taskDeadline)
		if er != nil {
			logger.Errorf("pull task status error: %v, retry(%d)", er, retryN)
			return true, er
		}

		// 正常情况下 pullTaskStepStatus() 应该在 runner 任务退出后才返回，
		// 但发现有任务在 running 状态时函数返回的情况，所以这里进行一次状态检查，如果任务不是退出状态则继续重试
		if !(models.TaskStep{}).IsExitedStatus(stepResult.Status) {
			logger.Warnf("pull task status done, but task status is '%s', retry(%d)", stepResult.Status, retryN)
			return true, fmt.Errorf("unexpected task step staus '%s'", stepResult.Status)
		}
		return false, nil
	})
	if err != nil {
		return stepResult, err
	}

	saveTaskStepResultFiles(task, step, stepResult.Result)

	message := ""
	switch stepResult.Status {
	case models.TaskStepFailed:
		message = "failed"
	case models.TaskStepTimeout:
		message = "timeout"
	case models.TaskStepAborted:
		message = "aborted"
	}

	if er := services.ChangeTaskStepStatusAndExitCode(
		sess, task, step, stepResult.Status, message, stepResult.Result.ExitCode); er != nil {
		return stepResult, er
	}
	return stepResult, err
}

func saveTaskStepResultFiles(task *models.Task, step *models.TaskStep, result runner.TaskStatusMessage) {
	logger := logs.Get().
		WithField("func", "saveTaskStepResultFiles").
		WithField("taskId", task.Id).
		WithField("step", fmt.Sprintf("%d(%s)", step.Index, step.Name))

	if len(result.LogContent) > 0 {
		content := result.LogContent
		content = logstorage.CutLogContent(content)
		if err := logstorage.Get().Write(step.LogPath, content); err != nil {
			logger.WithField("path", step.LogPath).Errorf("write task log error: %v", err)
			logger.Infof("task log content: %s", content)
		}
	}
	if len(result.TfStateJson) > 0 {
		path := task.StateJsonPath()
		if err := logstorage.Get().Write(path, result.TfStateJson); err != nil {
			logger.WithField("path", path).Errorf("write task state json error: %v", err)
		}
	}
	if len(result.TFProviderSchemaJson) > 0 {
		path := task.ProviderSchemaJsonPath()
		if err := logstorage.Get().Write(path, result.TFProviderSchemaJson); err != nil {
			logger.WithField("path", path).Errorf("write task provider json error: %v", err)
		}
	}
	if len(result.TfPlanJson) > 0 {
		path := task.PlanJsonPath()
		if err := logstorage.Get().Write(path, result.TfPlanJson); err != nil {
			logger.WithField("path", path).Errorf("write task plan json error: %v", err)
		}
	}
	if len(result.TfScanJson) > 0 {
		path := task.TfParseJsonPath()
		if err := logstorage.Get().Write(path, result.TfScanJson); err != nil {
			logger.WithField("path", path).Errorf("write task parse json error: %v", err)
		}
	}
	if len(result.TfResultJson) > 0 {
		path := task.TfResultJsonPath()
		if err := logstorage.Get().Write(path, result.TfResultJson); err != nil {
			logger.WithField("path", path).Errorf("write task scan result json error: %v", err)
		}
	}
}

func newReadMessageErr(err error) error {
	return errors.Wrap(err, "read message error")
}

// pullTaskStepStatus 获取任务最新状态，直到任务结束(或 ctx cancel)
// 该函数允许重复调用，即使任务己结束 (runner 会在本地保存近期(约7天)任务执行信息)，如果任务结束则写入全量日志到存储
func pullTaskStepStatus(ctx context.Context, task models.Tasker, step *models.TaskStep, deadline time.Time) (
	stepResult *waitStepResult, err error) {
	logger := logs.Get().WithField("action", "PullTaskState").WithField("taskId", task.GetId())

	runnerAddr, err := services.GetRunnerAddress(task.GetRunnerId())
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("envId", string(step.EnvId))
	params.Add("taskId", string(step.TaskId))
	params.Add("step", fmt.Sprintf("%d", step.Index))
	wsConn, resp, err := utils.WebsocketDail(runnerAddr, consts.RunnerTaskStepStatusURL, params)
	if err != nil {
		logger.Errorf("connect error: %v", err)
		if resp != nil && resp.StatusCode >= 300 {
			// 返回异常 http 状态码时表示请求参数有问题或者 runner 无法处理该连接，所以直接返回步骤失败
			return &waitStepResult{Status: models.TaskStepFailed, Result: runner.TaskStatusMessage{
				Exited:   true,
				ExitCode: 1,
			}}, nil
		}
		return stepResult, err
	}
	defer utils.WebsocketClose(wsConn)

	// 退出通知
	doneChan := make(chan struct{})
	defer close(doneChan)

	// 子协程在往 channel 写数据前先调用该函数检查主函数是否退出，若己退出则立即退出执行
	checkDone := func() bool {
		select {
		case <-doneChan:
			return true
		default:
			return false
		}
	}

	messageChan := make(chan *runner.TaskStatusMessage, 1)
	readErrChan := make(chan error, 1)

	readMessage := func() {
		defer close(messageChan)

		for {
			message := runner.TaskStatusMessage{}
			if err := wsConn.ReadJSON(&message); err != nil {
				if websocket.IsCloseError(err,
					websocket.CloseNormalClosure,
					websocket.CloseInternalServerErr) {
					logger.Traceln(newReadMessageErr(err))
				} else {
					logger.Warnln(newReadMessageErr(err))
					if checkDone() {
						return
					}
					readErrChan <- err
				}
				break
			} else {
				if checkDone() {
					return
				}
				messageChan <- &message
			}
		}
	}
	go readMessage()

	logger.Debugf("pulling step status, step=%s(%d)", step.Type, step.Index)
	stepResult, err = pullTaskStepStatusLoop(ctx, messageChan, readErrChan, deadline)
	if err != nil {
		return stepResult, err
	}
	logger.Debugf("pull step status done, step=%s(%d), status=%v code=%d",
		step.Type, step.Index, stepResult.Status, stepResult.Result.ExitCode)
	return stepResult, nil
}

func pullTaskStepStatusLoop(
	ctx context.Context,
	messageChan chan *runner.TaskStatusMessage,
	readErrChan chan error,
	deadline time.Time) (result *waitStepResult, err error) {

	now := time.Now()
	var timeout *time.Timer
	if deadline.Before(now) {
		// 即使任务己超时也保证进行一次状态获取
		timeout = time.NewTimer(time.Second)
	} else {
		timeout = time.NewTimer(deadline.Sub(now))
	}

	result = &waitStepResult{}
	for {
		select {
		case msg := <-messageChan:
			if msg == nil { // closed
				return result, nil
			}

			result.Result = *msg

			if msg.Timeout {
				result.Status = models.TaskStepTimeout
				return result, nil
			} else if msg.Aborted {
				result.Status = models.TaskStepAborted
				return result, nil
			} else if msg.Exited {
				if msg.ExitCode == 0 {
					result.Status = models.TaskStepComplete
				} else {
					result.Status = models.TaskStepFailed
				}
				return result, nil
			}
		case err := <-readErrChan:
			return result, newReadMessageErr(err)

		case <-ctx.Done():
			// logger.Infof("context done with: %v", ctx.Err())
			return result, nil

		case <-timeout.C:
			result.Status = models.TaskStepTimeout
			return result, nil
		}
	}
}

var (
	ErrTaskStepRejected = fmt.Errorf("rejected")
	ErrTaskStepAborted  = fmt.Errorf("aborted")
)

// WaitTaskStepApprove
// TODO: 使用注册通知机制，统一由一个 worker 来加载所有待审批的步骤最新状态，当有步骤审批通过时触发通知
func WaitTaskStepApprove(ctx context.Context, dbSess *db.Session, taskId models.Id, step int) (
	taskStep *models.TaskStep, err error) {

	ticker := time.NewTicker(consts.DbTaskPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			taskStep, err = services.GetTaskStep(dbSess, taskId, step)
			if err != nil {
				return nil, err
			}

			if taskStep.Status == models.TaskStepRejected {
				return nil, ErrTaskStepRejected
			} else if taskStep.Status == models.TaskStepAborted {
				return nil, ErrTaskStepAborted
			} else if taskStep.IsApproved() {
				return taskStep, nil
			}
		}
	}
}

// ========================================================
// 扫描任务

// WaitScanTaskStep 等待任务结束(包括超时)，返回任务最新状态
func WaitScanTaskStep(ctx context.Context, sess *db.Session, task *models.ScanTask, step *models.TaskStep) (
	stepResult *waitStepResult, err error) {

	logger := logs.Get().WithField("action", "WaitTaskStep").WithField("taskId", task.Id)
	if step.StartAt == nil {
		return nil, fmt.Errorf("step not start")
	}

	// runner 端己经增加了超时处理，portal 端的超时暂时保留，但时间设置为给定时间的 2 倍
	taskDeadline := time.Time(*step.StartAt).Add(time.Duration(task.StepTimeout*2) * time.Second)

	// 当前版本实现中需要 portal 主动连接到 runner 获取状态
	err = utils.RetryFunc(10, time.Second*5, func(retryN int) (retry bool, er error) {
		stepResult, er = pullTaskStepStatus(ctx, task, step, taskDeadline)
		if er != nil {
			logger.Errorf("pull task status error: %v, retry(%d)", er, retryN)
			return true, er
		}

		// 正常情况下 pullTaskStepStatus() 应该在 runner 任务退出后才返回，
		// 但发现有任务在 running 状态时函数返回的情况，所以这里进行一次状态检查，如果任务不是退出状态则继续重试
		if !(models.Task{}).IsExitedStatus(stepResult.Status) {
			logger.Warnf("pull task status done, but task status is '%s', retry(%d)", stepResult.Status, retryN)
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return stepResult, err
	}

	if len(stepResult.Result.LogContent) > 0 {
		content := stepResult.Result.LogContent
		content = logstorage.CutLogContent(content)
		if err := logstorage.Get().Write(step.LogPath, content); err != nil {
			logger.WithField("path", step.LogPath).Errorf("write task log error: %v", err)
			logger.Infof("task log content: %s", content)
		}
	}
	if len(stepResult.Result.TfScanJson) > 0 {
		path := task.TfParseJsonPath()
		if err := logstorage.Get().Write(path, stepResult.Result.TfScanJson); err != nil {
			logger.WithField("path", path).Errorf("write task parse json error: %v", err)
		}
	}
	if len(stepResult.Result.TfResultJson) > 0 {
		path := task.TfResultJsonPath()
		if err := logstorage.Get().Write(path, stepResult.Result.TfResultJson); err != nil {
			logger.WithField("path", path).Errorf("write task scan result json error: %v", err)
		}
	}
	// 合规任务暂时不需要发送消息
	//if stepResult.Status != models.TaskRunning && task.Extra.Source == consts.WorkFlow {
	//	k := kafka.Get()
	//	if err := k.ConnAndSend(k.GenerateKafkaContent(task.Extra.TransitionId, stepResult.Status)); err != nil {
	//		logger.Errorf("kafka send error: %v", err)
	//	}
	//}

	if er := services.ChangeTaskStepStatusAndExitCode(
		sess, task, step, stepResult.Status, "", stepResult.Result.ExitCode); er != nil {
		return stepResult, er
	}
	return stepResult, err
}
