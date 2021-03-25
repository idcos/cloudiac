package apps

import (
	"cloudiac/configs"
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/utils"
	"encoding/json"
	"fmt"
	"os"
)

type TaskDetailResp struct {
	models.Task
	models.Template
}

func TaskDetail(c *ctx.ServiceCtx, form *forms.DetailTaskForm) (interface{}, e.Error) {
	resp := TaskDetailResp{}
	if err := services.TaskDetail(c.DB().Debug(), form.TaskId).
		First(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return resp, nil
}

func TaskCreate(c *ctx.ServiceCtx, form *forms.CreateTaskForm) (interface{}, e.Error) {
	guid := utils.GenGuid("task")
	conf := configs.Get()
	logPath := fmt.Sprintf("%s/%s/%s", conf.Task.LogPath, form.TemplateGuid, guid)
	b, _ := json.Marshal(map[string]interface{}{
		"backend_url": fmt.Sprintf("http://%s:%s/api/v1", form.RunnerIp, form.RunnerPort),
		"log_file":    logPath,
		"log_offset":  0,
	})

	if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
		return nil, e.New(e.IOError, err)
	}

	return services.CreateTask(c.DB().Debug(), models.Task{
		TemplateId:   form.TemplateId,
		TemplateGuid: form.TemplateGuid,
		Guid:         guid,
		TaskType:     form.TaskType,
		Status:       consts.TaskPending,
		Creator:      form.Creator,
		TaskName:     form.TaskName,
		BackendInfo:  models.JSON(b),
	})
}
