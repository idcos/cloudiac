package apps

import (
	"bufio"
	"bytes"
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models"
	"cloudiac/services"
	"cloudiac/services/logstorage"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"fmt"
	"github.com/gin-contrib/sse"
	"github.com/gorilla/websocket"
	"io"
	"net/url"
	"strconv"
	"strings"
)

func FollowTaskLog(c *ctx.GinRequestCtx) error {
	logPath := c.Query("logPath")
	// logPath example: "logs/ct-c2j2g5rn8qhqp9ku9a6g/run-c2mdu4ecie6qs8gmsmkgg"
	parts := strings.Split(logPath, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid log path: '%v'", logPath)
	}

	taskId := parts[len(parts)-1]
	logger := logs.Get().WithField("func", "FollowTaskLog").WithField("taskId", taskId)

	task, err := services.GetTaskByGuid(c.ServiceCtx().DB(), taskId)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return e.New(e.TaskNotExists)
		}
		logger.Errorf("query task err: %v", err)
		return e.New(e.DBError)
	}

	var reader io.Reader

	if task.Exited() { // 己退出的任务直接读取全量日志
		if content, err := logstorage.Get().Read(task.FullLogPath()); err != nil {
			logger.Errorf("read task log error: %v", err)
			return err
		} else {
			reader = bytes.NewBuffer(content)
		}
	} else { // 否则实时从 runner 获取日志
		pr, pw := io.Pipe()
		reader = pr

		go func() {
			if err := fetchRunnerTaskLog(pw, task); err != nil {
				logger.Errorf("fetchRunnerTaskLog error: %v", err)
			}
		}()

		go func() {
			<-c.Request.Context().Done()
			logger.Tracef("connect closed")
			pr.Close()
		}()
	}

	scanner := bufio.NewScanner(reader)
	eventId := 0 // to indicate the message id
	for scanner.Scan() {
		c.Render(-1, sse.Event{
			Id:    strconv.Itoa(eventId),
			Event: "message",
			Data:  scanner.Text(),
		})
		c.Writer.Flush()
		eventId += 1
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return err
	}
	return nil
}

// 从 runner 获取任务日志，直到任务结束
func fetchRunnerTaskLog(writer io.WriteCloser, task *models.Task) error {
	// close 后 read 端会触发 EOF error
	defer writer.Close()

	logger := logs.Get().WithField("func", "fetchRunnerTaskLog").WithField("taskId", task.Guid)

	taskBackend := task.UnmarshalBackendInfo()
	runnerAddr := fmt.Sprintf("%v", taskBackend.BackendUrl)

	params := url.Values{}
	params.Add("taskId", task.Guid)
	params.Add("templateId", task.TemplateGuid)
	params.Add("containerId", task.UnmarshalBackendInfo().ContainerId)
	wsConn, err := utils.WebsocketDail(runnerAddr, consts.RunnerTaskLogFollowURL, params)
	if err != nil {
		return err
	}
	defer func() {
		_ = utils.WebsocketClose(wsConn)
	}()

	for {
		_, reader, err := wsConn.NextReader()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				logger.Tracef("read message error: %v", err)
				return nil
			} else {
				logger.Errorf("read message error: %v", err)
				return err
			}
		} else {
			_, err := io.Copy(writer, reader)
			if err != nil {
				logger.Debugf("io.Copy: %v", err)
				if err == io.ErrClosedPipe {
					return nil
				}
				return err
			}
		}
	}
}
