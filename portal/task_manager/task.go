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
	"cloudiac/utils/kafka"
	"cloudiac/utils/logs"
)

// StartTaskStep 启动任务的一步
// 该函数会设置 taskReq 中 step 相关的数据
func StartTaskStep(taskReq runner.RunTaskReq, step models.TaskStep) (err error) {
	logger := logs.Get().
		WithField("action", "StartTaskStep").
		WithField("taskId", taskReq.TaskId).
		WithField("step", step.Index)

	header := &http.Header{}
	header.Set("Content-Type", "application/json")

	var runnerAddr string
	runnerAddr, err = services.GetRunnerAddress(taskReq.RunnerId)
	if err != nil {
		return errors.Wrapf(err, "get runner '%s' address", taskReq.RunnerId)
	}

	requestUrl := utils.JoinURL(runnerAddr, consts.RunnerRunTaskURL)
	logger.Debugf("request runner: %s", requestUrl)

	taskReq.Step = step.Index
	taskReq.StepType = step.Type
	taskReq.StepArgs = step.Args

	respData, err := utils.HttpService(requestUrl, "POST", header, taskReq, 1, 5)
	if err != nil {
		return err
	}

	resp := runner.Response{}
	if err := json.Unmarshal(respData, &resp); err != nil {
		return fmt.Errorf("unexpected response: %s", respData)
	}
	logger.Debugf("runner response: %#v", resp)
	if resp.Error != "" {
		return fmt.Errorf(resp.Error)
	}
	return nil
}

type waitStepResult struct {
	Status string
	Result runner.TaskStatusMessage
}

// WaitTaskStep 等待任务结束(包括超时)，返回任务最新状态
// 该函数会更新任务状态、日志等到 db
// param: taskDeadline 任务超时时间，达到这个时间后任务会被置为 timeout 状态
func WaitTaskStep(ctx context.Context, sess *db.Session, task *models.Task, step *models.TaskStep) (
	stepResult *waitStepResult, err error) {

	logger := logs.Get().WithField("action", "WaitTaskStep").WithField("taskId", task.Id)
	if step.StartAt == nil {
		return nil, fmt.Errorf("step not start")
	}
	taskDeadline := time.Time(*step.StartAt).Add(time.Duration(task.StepTimeout) * time.Second)

	// 当前版本实现中需要 portal 主动连接到 runner 获取状态
	err = utils.RetryFunc(0, time.Second*10, func(retryN int) (retry bool, er error) {
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

	if stepResult.Status != models.TaskRunning && task.Extra.Source == consts.WorkFlow {
		k := kafka.Get()
		if err := k.ConnAndSend(k.GenerateKafkaContent(task.Extra.TransitionId, stepResult.Status)); err != nil {
			logger.Errorf("kafka send error: %v", err)
		}
	}

	if len(stepResult.Result.LogContent) > 0 {
		content := stepResult.Result.LogContent
		content = logstorage.CutLogContent(content)
		if err := logstorage.Get().Write(step.LogPath, content); err != nil {
			logger.WithField("path", step.LogPath).Errorf("write task log error: %v", err)
			logger.Infof("task log content: %s", content)
		}
	}
	if len(stepResult.Result.TfStateJson) > 0 {
		path := task.StateJsonPath()
		if err := logstorage.Get().Write(path, stepResult.Result.TfStateJson); err != nil {
			logger.WithField("path", path).Errorf("write task state json error: %v", err)
		}
	}
	if len(stepResult.Result.TfPlanJson) > 0 {
		path := task.PlanJsonPath()
		if err := logstorage.Get().Write(path, stepResult.Result.TfPlanJson); err != nil {
			logger.WithField("path", path).Errorf("write task plan json error: %v", err)
		}
	}

	if er := services.ChangeTaskStepStatus(sess, task, step, stepResult.Status, ""); er != nil {
		return stepResult, er
	}
	return stepResult, err
}

// pullTaskStepStatus 获取任务最新状态，直到任务结束(或 ctx cancel)
// 该函数允许重复调用，即使任务己结束 (runner 会在本地保存近期(约7天)任务执行信息)，如果任务结束则写入全量日志到存储
func pullTaskStepStatus(ctx context.Context, task *models.Task, step *models.TaskStep, deadline time.Time) (
	stepResult *waitStepResult, err error) {
	logger := logs.Get().WithField("action", "PullTaskState").WithField("taskId", task.Id)

	runnerAddr, err := services.GetRunnerAddress(task.RunnerId)
	if err != nil {
		return nil, errors.Wrapf(err, "get runner address")
	}

	params := url.Values{}
	params.Add("envId", string(task.EnvId))
	params.Add("taskId", string(task.Id))
	params.Add("step", fmt.Sprintf("%d", step.Index))
	wsConn, resp, err := utils.WebsocketDail(runnerAddr, consts.RunnerTaskStateURL, params)
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
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					logger.Tracef("read message error: %v", err)
				} else {
					logger.Errorf("read message error: %v", err)
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

	now := time.Now()
	var timeout *time.Timer
	if deadline.Before(now) {
		// 即使任务己超时也保证进行一次状态获取
		timeout = time.NewTimer(time.Second)
	} else {
		timeout = time.NewTimer(deadline.Sub(now))
	}

	//var lastStatus *runner.TaskStatusMessage
	stepResult = &waitStepResult{}
	selectLoop := func() error {
		for {
			select {
			case msg := <-messageChan:
				if msg == nil { // closed
					return nil
				}

				stepResult.Result = *msg
				//logger.Debugf("receive task status message: %v, %v", msg.Exited, msg.ExitCode)
				if msg.Exited {
					if msg.ExitCode == 0 {
						stepResult.Status = models.TaskComplete
					} else {
						stepResult.Status = models.TaskFailed
					}
					return nil
				}
			case err = <-readErrChan:
				return fmt.Errorf("read message error: %v", err)

			case <-ctx.Done():
				logger.Infof("context done with: %v", ctx.Err())
				return nil

			case <-timeout.C:
				stepResult.Status = models.TaskStepTimeout
				return nil
			}
		}
	}

	logger.Infof("pulling step status ...")
	err = selectLoop()
	logger.Infof("pull step status done, status=%v", stepResult.Status)

	return stepResult, nil
}

var (
	ErrTaskStepRejected = fmt.Errorf("rejected")
)

// WaitTaskStepApprove
// TODO: 使用注册通知机制，统一由一个 worker 来加载所有待审批的步骤最新状态，当有步骤审批通过时触发通知
func WaitTaskStepApprove(ctx context.Context, dbSess *db.Session, taskId models.Id, step int) (
	taskStep *models.TaskStep, err error) {

	ticker := time.NewTicker(time.Second * 5)
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
			} else if taskStep.IsApproved() {
				return taskStep, nil
			}
		}
	}
}
