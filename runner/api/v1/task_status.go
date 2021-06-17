package v1

import (
	"cloudiac/runner"
	"cloudiac/runner/ws"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var logger = logs.Get()

func TaskStatus(c *gin.Context) {
	task := runner.CommitedTask{
		TemplateId:  c.Query("templateId"),
		TaskId:      c.Query("taskId"),
		ContainerId: c.Query("containerId"),
	}

	logger := logger.WithField("taskId", task.TaskId)
	wsConn, peerClosed, err := ws.UpgradeWithNotifyClosed(c.Writer, c.Request, nil)
	if err != nil {
		logger.Warnln(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer func() {
		wsConn.Close()
	}()

	if err := doTaskStatus(wsConn, &task, peerClosed); err != nil {
		logger.Errorln(err)
		utils.WebsocketCloseWithCode(wsConn, websocket.CloseInternalServerErr, err.Error())
	} else {
		utils.WebsocketClose(wsConn)
	}
}

func doTaskStatus(wsConn *websocket.Conn, task *runner.CommitedTask, closed <-chan struct{}) error {
	logger := logger.WithField("taskId", task.TaskId)

	// 获取任务最新状态并通过 websocket 发送
	sendStatus := func(withLog bool) error {
		containerJSON, err := task.Status()
		if err != nil {
			return err
		}

		msg := runner.TaskStatusMessage{
			Status:     containerJSON.State.Status,
			StatusCode: containerJSON.State.ExitCode,
		}

		if withLog {
			logs, err := runner.FetchTaskLog(task.TemplateId, task.TaskId)
			if err != nil {
				logger.Errorf("fetch task log error: %v", err)
				msg.LogContent = utils.TaskLogMsgBytes("Fetch task log error: %v", err)
			} else {
				msg.LogContent = logs
			}

			stateList, err := runner.FetchStateList(task.TemplateId, task.TaskId)
			if err != nil {
				logger.Errorf("fetch task log error: %v", err)
				msg.StateListContent = utils.TaskLogMsgBytes("Fetch state list error: %v", err)
			} else {
				msg.StateListContent = stateList
			}

		}

		if err := wsConn.WriteJSON(msg); err != nil {
			logger.Errorln(err)
			return err
		}
		return nil
	}

	ctx, cancelFun := context.WithCancel(context.Background())
	defer cancelFun()

	waitCh := make(chan error, 1)
	go func() {
		defer close(waitCh)

		_, err := task.Wait(ctx)
		waitCh <- err
	}()

	if err := sendStatus(false); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	logger.Infof("watching task status")
	defer logger.Infof("watch task status done")
	for {
		select {
		case <-closed:
			logger.Debugf("connection closed")
			cancelFun()
		case <-ticker.C:
			// 定时发送最新任务状态
			if err := sendStatus(false); err != nil {
				logger.Warnf("send status error: %v", err)
			}
		case err := <-waitCh:
			if ctx.Err() != nil { // 对端断开连接
				return nil
			}
			if err != nil {
				return err
			}
			// 任务结束，发送状态和全量日志
			return sendStatus(true)
		}
	}
}
