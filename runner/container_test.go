package runner

import (
	"log"
	"time"
)

func ExampleContainerWait() {
	task := CommitedTask{
		TemplateId:       "tplId",
		TaskId:           "taskId",
		ContainerId:      "",
		LogContentOffset: 0,
	}

	_, err := task.Wait(time.Second)
	if err != nil {
		log.Fatal(err)
	}
}
