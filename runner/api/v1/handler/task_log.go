package handler

import (
	"bufio"
	"cloudiac/runner"
	"cloudiac/runner/api/ctx"
	"cloudiac/runner/ws"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// TaskLogFollow 读取 task log 并 follow, 直到任务退出
func TaskLogFollow(c *ctx.Context) {
	req := runner.TaskLogReq{}
	if err := c.BindQuery(&req); err != nil {
		c.Error(err, http.StatusBadRequest)
		return
	}

	task, err := runner.LoadCommittedTask(req.EnvId, req.TaskId, req.Step)
	if err != nil {
		if os.IsNotExist(err) {
			c.Error(err, http.StatusNotFound)
		} else {
			c.Error(err, http.StatusInternalServerError)
		}
		return
	}

	logger := logger.WithField("taskId", task.TaskId)
	wsConn, peerClosed, err := ws.UpgradeWithNotifyClosed(c.Writer, c.Request, nil)
	if err != nil {
		logger.Warnln(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer utils.WebsocketClose(wsConn)

	if err := doFollowTaskLog(wsConn, task, 0, peerClosed); err != nil {
		logger.Errorf("doFollowTaskLog error: %v", err)
		_ = utils.WebsocketCloseWithCode(wsConn, websocket.CloseInternalServerErr, err.Error())
	} else {
		_ = utils.WebsocketClose(wsConn)
	}
}

func doFollowTaskLog(wsConn *websocket.Conn, task *runner.CommittedTaskStep, offset int64, closedCh <-chan struct{}) error {
	logger := logger.WithField("func", "doFollowTaskLog").WithField("taskId", task.TaskId)

	var (
		taskExitChan = make(chan error)
	)

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	logPath := filepath.Join(runner.GetTaskStepDir(task.EnvId, task.TaskId, task.Step), runner.TaskLogName)
	contentChan, readErrChan := followFile(ctx, logPath, offset)

	// 等待任务退出协程
	go func() {
		defer close(taskExitChan)

		_, err := task.Wait(ctx)
		// followFile() 会在遇到 EOF 时延迟一定时间进行下一次读取，
		// 如果这里立即发送信号 follow 会在当延迟结束后立即退出，导致最后写入的日志没有被读取
		time.Sleep(runner.FollowLogDelay)
		taskExitChan <- err
	}()

	for {
		select {
		case content := <-contentChan:
			if err := wsConn.WriteMessage(websocket.TextMessage, content); err != nil {
				logger.Errorf("write message error: %v", err)
				return err
			}
		case err := <-readErrChan:
			if err != nil {
				logger.Errorf("read content error: %v", err)
				return err
			}
		case err := <-taskExitChan:
			if err != nil {
				logger.Errorf("wait task error: %v", err)
			} else {
				logger.Infof("task finshed")
			}
			return err
		case <-closedCh:
			logger.Debugf("connection closedCh")
			return nil
		}
	}
}

// 读取文件内容并 follow，直到 ctx 被 cancel
// return: 两个 chan，一个用于返回文件内容，一个用于返回 err，chan 在函数退出时会被关闭，所以会读到 nil
func followFile(ctx context.Context, path string, offset int64) (<-chan []byte, <-chan error) {
	logger := logs.Get().WithField("func", "followFile").WithField("path", path)

	var (
		contentChan = make(chan []byte)
		errChan     = make(chan error, 1)
	)

	logFp, err := os.Open(path)
	if err != nil {
		errChan <- err
		return contentChan, errChan
	}

	if offset != 0 {
		if _, err := logFp.Seek(offset, 0); err != nil {
			_ = logFp.Close()
			return contentChan, errChan
		}
	}

	go func() {
		defer func() {
			_ = logFp.Close()
			close(contentChan)
			close(errChan)
		}()

		reader := bufio.NewReader(logFp)
		for {
			content, err := reader.ReadBytes('\n')
			if len(content) > 0 {
				contentChan <- content
			}

			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					if err == io.EOF {
						// 读到了文件末尾，等待一下再进行下一次读取
						time.Sleep(runner.FollowLogDelay)
						continue
					} else {
						errChan <- err
						return
					}
				}
			}
		}
	}()

	go func() {
		<-ctx.Done()
		logger.Tracef("context done, %v", ctx.Err())
		// 关闭文件，中断 Read()
		_ = logFp.Close()
	}()

	return contentChan, errChan
}
