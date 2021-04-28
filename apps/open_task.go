package apps

import (
	"cloudiac/libs/ctx"
	"cloudiac/services"
	"cloudiac/utils/logs"
	"encoding/json"
)

func TaskLogSSEGetPath(c *ctx.ServiceCtx, taskGuid string) string {
	task, err := services.GetTaskByGuid(c.DB(),taskGuid)
	if err != nil {
		logs.Get().Error(err)
	}
	taskBackend := make(map[string]interface{}, 0)
	json.Unmarshal(task.BackendInfo, &taskBackend)
	return taskBackend["log_file"].(string)
}
