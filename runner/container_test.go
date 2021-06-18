package runner

import (
	"context"
	"log"
	"time"
)

func ExampleContainerWait() {
	task := CommitedTask{
		TemplateId:       "tplId",
		TaskId:           "taskId",
		ContainerId:      "",
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	_, err := task.Wait(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
