package runner

import (
	"net/http"

	"github.com/docker/docker/api/types"
)

type ContainerStatus struct {
	Status          *types.ContainerState
	LogContent      []string
	LogContentLines int
}

func Run(req *http.Request) (string, error) {
	c, state, iacTemplate, err := ReqToCommand(req)
	if err != nil {
		return "", err
	}

	// TODO(ZhengYUe):
	// 1. 根据模板ID创建目录(目录创建规则：/{template_uuid}/{task_id}/)，用于保存日志文件及挂载provider、密钥等文件
	// 2. 若需要保存模板状态，则根据参数生成状态配置文件，放入模板目录中，挂载至容器内部

	templateUUID := iacTemplate.TemplateUUID
	taskId := iacTemplate.TaskId
	templateDir, err := CreateTemplatePath(templateUUID, taskId)
	if err != nil {
		return "", err
	}

	// if state.SaveState != false {
	GenStateFile(state.StateBackendAddress, state.Schema, state.StateKey, templateDir)
	// }
	err = c.Create(templateDir)
	return c.ContainerInstance.ID, err
}

func Cancel(req *http.Request) error {
	task, err := ReqToTask(req)
	if err != nil {
		return err
	}
	err = task.Cancel()
	return err
}

func Status(req *http.Request) (ContainerStatus, error) {
	task, err := ReqToTask(req)
	containerStatus := new(ContainerStatus)
	if err != nil {
		return *containerStatus, err
	}
	containerJSON, err := task.Status()
	if err != nil {
		return *containerStatus, err
	}

	containerStatus.Status = containerJSON.State

	logContent, err := FetchTaskLog(task.TemplateUUID, task.TaskId, task.LogContentOffset)
	if err != nil {
		return *containerStatus, err
	}
	containerStatus.LogContentLines = len(logContent)
	containerStatus.LogContent = logContent
	return *containerStatus, nil
}
