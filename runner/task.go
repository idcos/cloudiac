package runner

import (
	"github.com/pkg/errors"
	"net/http"

	"github.com/docker/docker/api/types"
)

type ContainerStatus struct {
	Status     *types.ContainerState
	LogContent []string
}

func Run(req *http.Request) (string, error) {
	c, state, err := ReqToCommand(req)
	if err != nil {
		return "", err
	}

	if err := GenInjectTfConfig(InjectConfigContext{
		WorkDir:          c.TaskWorkdir,
		BackendAddress:   state.StateBackendAddress,
		BackendScheme:    state.Scheme,
		BackendPath:      state.StateKey,
	}, c.PrivateKey); err != nil {
		return "", errors.Wrap(err, "generate inject config error")
	}
	if err = c.Create(); err != nil {
		return "", err
	}
	return c.ContainerInstance.ID, nil
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
