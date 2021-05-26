package apps

import (
	"bufio"
	"cloudiac/consts"
	"cloudiac/libs/ctx"
	"cloudiac/libs/db"
	"cloudiac/services"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"encoding/json"
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

	pr, pw := io.Pipe()
	go func() {
		if err := fetchRunnerTaskLog(c.ServiceCtx().DB(), pw, taskId); err != nil {
			logger.Errorf("fetchRunnerTaskLog error: %v", err)
		}
	}()

	go func() {
		<- c.Request.Context().Done()
		logger.Debugf("connect closed")
		pr.Close()
	}()

	scanner := bufio.NewScanner(pr)
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
func fetchRunnerTaskLog(dbSess *db.Session, writer io.WriteCloser, taskId string) error {
	// close 后 read 端会触发 EOF error
	defer writer.Close()

	logger := logs.Get().WithField("func", "fetchRunnerTaskLog").WithField("taskId", taskId)

	task, er := services.GetTaskByGuid(dbSess, taskId)
	if er != nil {
		logger.Errorf("query task err: %v", er)
		return er
	}

	taskBackend := make(map[string]interface{})
	_ = json.Unmarshal(task.BackendInfo, &taskBackend)
	runnerAddr := fmt.Sprintf("%v", taskBackend["backend_url"])

	params := url.Values{}
	params.Add("taskId", taskId)
	params.Add("templateId", task.TemplateGuid)
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
			io.Copy(writer, reader)
		}
	}
}
