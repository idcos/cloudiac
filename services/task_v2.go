package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
	"time"

	"cloudiac/consts"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/runner"
	"cloudiac/utils"
	"cloudiac/utils/kafka"
	"cloudiac/utils/logs"
)

// WaitTask 等待任务结束(或超时)，返回任务最新状态
func WaitTask(ctx context.Context, taskId string, tpl *models.Template, dbSess *db.Session) (status string, err error) {
	logger := logs.Get().WithField("action", "WaitTask").WithField("taskId", taskId)
	retryCount := 0
	maxRetry := 10	// 最大重试次数(不含第一次)
	for {
		status, err = doPullTaskStatus(ctx, taskId, tpl, dbSess)
		if err != nil {
			retryCount += 1
			if retryCount > maxRetry {
				return status, err
			}

			sleepTime := retryCount * 2
			if sleepTime > 10 {
				sleepTime = 10
			}
			logger.Errorf("pull task status error: %v, retry after %ds", err, sleepTime)
			time.Sleep(time.Duration(sleepTime) * time.Second)
			continue
		} else {
			return status, nil
		}
	}
}

// PullTaskStatus 同步任务最新状态，直到任务结束
// 该函数应该允许重复调用，即使任务己结束 (runner 会在本地保存近期(约7天)任务执行信息)
func doPullTaskStatus(ctx context.Context, taskId string, tpl *models.Template, dbSess *db.Session) (
	taskStatus string, err error) {
	logger := logs.Get().WithField("action", "PullTaskState").WithField("taskId", taskId)

	// 获取 task 最新状态
	task, err := GetTaskByGuid(dbSess, taskId)
	if err != nil {
		logger.Errorf("query task err: %v", err)
		return "", err
	}
	taskStatus = task.Status

	taskBackend := make(map[string]interface{}, 0)
	_ = json.Unmarshal(task.BackendInfo, &taskBackend)
	runnerAddr := taskBackend["backend_url"]

	params := url.Values{}
	params.Add("templateUuid", tpl.Guid)
	params.Add("taskId", task.Guid)
	params.Add("containerId", fmt.Sprintf("%s", taskBackend["container_id"]))
	wsConn, err := utils.WebsocketDail(fmt.Sprintf("%s", runnerAddr), consts.RunnerTaskStateURL, params)
	if err != nil {
		logger.Errorln(err)
		return taskStatus, err
	}
	defer func() {
		wsConn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(time.Second),
		)
		wsConn.Close()
	}()

	messageChan := make(chan *runner.TaskStatusMessage)
	readErrChan := make(chan error)
	readMessage := func() {
		defer close(messageChan)

		for {
			message := runner.TaskStatusMessage{}
			if err := wsConn.ReadJSON(&message); err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					logger.Tracef("read message error: %v", err)
				} else {
					logger.Errorf("read message error: %v", err)
					readErrChan <- err
				}
				break
			} else {
				messageChan <- &message
			}
		}
	}
	go readMessage()

	deadline := task.StartAt.Add(time.Duration(tpl.Timeout) * time.Second)
	now := time.Now()
	var timer *time.Timer
	if deadline.Before(now) {
		timer = time.NewTimer(time.Second * 10)
	} else {
		timer = time.NewTimer(deadline.Sub(now))
	}
	var lastMessage *runner.TaskStatusMessage

forLoop:
	for {
		select {
		case msg := <-messageChan:
			if msg == nil { // closed
				break forLoop
			}

			lastMessage = msg
			if lastMessage.Status == consts.DockerStatusExited {
				if msg.StatusCode == 0 {
					taskStatus = consts.TaskComplete
				} else {
					taskStatus = consts.TaskFailed
				}
				break
			}

		case err = <-readErrChan:
			return taskStatus, fmt.Errorf("read message error: %v", err)

		case <-ctx.Done():
			logger.Infof("context done with: %v", ctx.Err())
			return taskStatus, nil

		case <-timer.C:
			taskStatus = consts.TaskTimeout
			break forLoop
		}
	}

	if taskStatus != consts.TaskRunning && len(lastMessage.LogContent) > 0 {
		if err := writeTaskLog(lastMessage.LogContent,
			fmt.Sprintf("%s", taskBackend["log_file"]), 0); err != nil {
			logger.Errorf("write task log error: %v", err)
			logger.Infof("task log content: %v", lastMessage.LogContent)
		}
	}

	updateM := map[string]interface{}{
		"status": taskStatus,
		"end_at": time.Now(),
	}
	updateM["end_at"] = time.Now()
	if taskStatus != consts.TaskRunning && task.Source == consts.WorkFlow {
		k := kafka.Get()
		if err := k.ConnAndSend(k.GenerateKafkaContent(task.TransactionId, taskStatus)); err != nil {
			logger.Errorf("kafka send error: %v", err)
		}
	}

	//更新task状态
	if _, err := dbSess.Model(&models.Task{}).Where("id = ?", task.Id).Update(updateM); err != nil {
		return taskStatus, err
	}
	return taskStatus, nil
}
