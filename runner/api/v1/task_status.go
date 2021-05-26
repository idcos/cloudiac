package v1

import (
	"cloudiac/runner"
	"cloudiac/runner/ws"
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
	wsConn, err := ws.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Warnln(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sendCloseMessage := func(code int, text string) {
		message := websocket.FormatCloseMessage(code, "")
		wsConn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))
	}

	defer func() {
		wsConn.Close()
	}()

	// 通知处理进程，对端主动断开了连接
	peerClosed := make(chan struct{})
	go func() {
		for {
			// 通过调用 ReadMessage() 来检测对端是否断开连接，
			// 如果对端关闭连接，该调用会返回 error，其他消息我们忽略
			_, _, err := wsConn.ReadMessage()
			if err != nil {
				close(peerClosed)
				if !websocket.IsUnexpectedCloseError(err) {
					logger.Debugf("read ws message: %v", err)
				}
				return
			}
		}
	}()

	if err := doTaskStatus(wsConn, &task, peerClosed); err != nil {
		logger.Errorln(err)
		sendCloseMessage(websocket.CloseInternalServerErr, err.Error())
	} else {
		sendCloseMessage(websocket.CloseNormalClosure, "")
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
			logs, err := runner.FetchTaskLog(task.TemplateId, task.TaskId, 0)
			if err != nil {
				logger.Errorf("fetch task log error: %v", err)
			}
			msg.LogContent = logs
			msg.LogContentLines = len(logs)
		}

		if err := wsConn.WriteJSON(msg); err != nil {
			logger.Errorln(err)
			return err
		}
		return nil
	}

	waitCh := make(chan error)
	go func() {
		defer close(waitCh)

		_, err := task.Wait(context.Background())
		waitCh <- err
	}()

	if err := sendStatus(false); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	logger.Infof("watching task status")
	for {
		select {
		case <-closed:
			logger.Debugf("peer connection closed")
			return nil
		case err := <-waitCh:
			if err != nil {
				return err
			}
			// 任务结束，发送状态和全量日志
			return sendStatus(true)
		case <-ticker.C:
			// 定时发送最新任务状态
			if err := sendStatus(false); err != nil {
				return err
			}
		}
	}
}
