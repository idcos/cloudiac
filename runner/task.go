package runner

import (
	"github.com/pkg/errors"
	"net/http"

	"github.com/docker/docker/api/types"
)

type ContainerStatus struct {
	Status          *types.ContainerState
	LogContent      []string
}

func Run(req *http.Request) (string, error) {
	c, state, err := ReqToCommand(req)
	if err != nil {
		return "", err
	}

	// TODO(ZhengYue):
	// 1. 根据模板ID创建目录(目录创建规则：/{template_uuid}/{task_id}/)，用于保存日志文件及挂载provider、密钥等文件
	// 2. 若需要保存模板状态，则根据参数生成状态配置文件，放入模板目录中，挂载至容器内部

	if err := GenBackendConfig(state.StateBackendAddress, state.Scheme, state.StateKey, c.TaskWorkdir); err != nil {
		return "", errors.Wrap(err, "generate backend config error")
	}
	err = c.Create()
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

type TaskLogsResp struct {
	LogContent      []string
	LogContentLines int
}
