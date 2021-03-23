package handlers

import (
	"cloudiac/libs/ctx"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sse"
)

var old_size = map[string]int{}
var end_str = map[string]string{}

func getEvent(filename string) (*sse.Event, error) {
	if end_str[filename] == "end\n" {
		return nil, io.EOF
	}

	file, err := os.Open(fmt.Sprintf("/tmp/%s", filename))
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	fileinfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	fileSize := fileinfo.Size()
	buffer := make([]byte, fileSize)

	new_size, err := file.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("read %d %d bytes: %q\n", old_size[filename], new_size, buffer[:new_size])

	time.Sleep(time.Second)
	event := sse.Event{
		Event: "hello",
		Data:  fmt.Sprintf("%s", buffer[old_size[filename]:new_size]),
	}
	old_size[filename] = new_size

	end := new_size - 4
	if end > 0 {
		end_str[filename] = fmt.Sprintf("%s", buffer[end:new_size])
	}

	return &event, nil
}

func HelloSse(c *ctx.GinRequestCtx) {
	old_size[c.Param("filename")] = 0
	for {
		event, err := getEvent(c.Param("filename"))

		if err == io.EOF {
			c.Status(http.StatusNoContent)
			return
		}
		if event.Data == "" {
			continue
		}

		select {

		case <-c.Request.Context().Done():
			return

		default:
			_ = event.Render(c.Writer)
			c.Writer.Flush()
		}
	}
}