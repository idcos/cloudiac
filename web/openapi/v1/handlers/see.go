package handlers

import (
	"bufio"
	"cloudiac/apps"
	"cloudiac/consts"
	"cloudiac/libs/ctx"
	"cloudiac/utils/logs"
	"context"
	"fmt"
	"github.com/gin-contrib/sse"
	"io"
	"os"
	"strconv"
	"time"
)

func TaskLogSSE(c *ctx.GinRequestCtx) {
	loggers := logs.Get()
	contx, cancel := context.WithCancel(context.Background())
	defer cancel()
	chanStream := make(chan string, 0)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-c.Request.Context().Done():
				// client gave up
				done <- true
				return
			case <-contx.Done():
				switch contx.Err() {
				case context.DeadlineExceeded:
					loggers.Error("timeout")
				}
				done <- true
				return
			}
		}
	}()

	logPath := apps.TaskLogSSEGetPath(c.ServiceCtx(), c.Query("taskGuid"))

	path := fmt.Sprintf("%s/%s", logPath, consts.TaskLogName)
	f, err := os.Open(path)
	if err != nil {
		loggers.Error(err)
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	fi, err := f.Stat()
	if err != nil {
		loggers.Error(err)
		return
	}

	go func() {
		for {
			str, _, err := rd.ReadLine()
			if err != nil {
				if err.Error() == "EOF" {
					if time.Now().Unix()-fi.ModTime().Unix() > 60 {
						done <- true
						return
					}
					time.Sleep(1000)
					continue
				} else {
					loggers.Error("Read Error:", err.Error())
					done <- true
					return
				}
			}
			chanStream <- string(str)
		}
	}()

	count := 0 // to indicate the message id
	isStreaming := c.Stream(func(w io.Writer) bool {
		for {
			select {
			case <-done:
				// when deadline is reached, send 'end' event
				c.SSEvent("end", "end")
				return false
			case msg := <-chanStream:
				// send events to client
				c.Render(-1, sse.Event{
					Id:    strconv.Itoa(count),
					Event: "message",
					Data:  msg,
				})
				count++
				return true
			}
		}
	})
	if !isStreaming {
		loggers.Info("Stream Closed!")
	}
}
