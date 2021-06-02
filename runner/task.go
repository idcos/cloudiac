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
