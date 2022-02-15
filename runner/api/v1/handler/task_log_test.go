// Copyright 2021 CloudJ Company Limited. All rights reserved.

package handler

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

func TestExampleFollowFile(t *testing.T) {
	logger := log.Default()

	fp, err := ioutil.TempFile("", "")
	logger.Printf("file: %v", fp.Name())
	if err != nil {
		logger.Panic(err)
	}

	defer func() {
		_ = fp.Close()
		_ = os.Remove(fp.Name())
	}()

	go func() {
		for {
			_, err := fp.WriteString(time.Now().String() + "\n")
			if err != nil {
				logger.Panic(err)
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	contentC, errC := followFile(ctx, fp.Name(), 0)
	for {
		select {
		case content := <-contentC:
			if len(content) > 0 {
				logger.Printf("content: %s", content)
			}
		case err := <-errC:
			if err != nil {
				if err != context.DeadlineExceeded {
					logger.Fatalln(err)
				} else {
					logger.Println(err)
				}
			}
			return
		}
	}
}
