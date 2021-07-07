package runner

import (
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"os"
	"path/filepath"
	"sync"
)

var logger = logs.Get()

type CommittedTaskStep struct {
	EnvId       string     `json:"envId"`
	TaskId      string     `json:"taskId"`
	Step        int        `json:"step"`
	Request     RunTaskReq `json:"request"`
	ContainerId string     `json:"containerId"`

	containerInfoLock sync.RWMutex
}

func LoadCommittedTask(envId string, taskId string, step int) (*CommittedTaskStep, error) {
	path := filepath.Join(GetTaskStepDir(envId, taskId, step), TaskInfoFileName)
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	task := CommittedTaskStep{}
	if err := json.NewDecoder(fp).Decode(&task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (task *CommittedTaskStep) Cancel() error {
	cli, err := client.NewClientWithOpts()
	cli.NegotiateAPIVersion(context.Background())
	if err != nil {
		logger.Warnf("unable to create docker client, error: %v", err)
		return err
	}

	containerRemoveOpts := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	if err := cli.ContainerRemove(context.Background(), task.ContainerId, containerRemoveOpts); err != nil {
		if _, ok := err.(errdefs.ErrNotFound); ok {
			return nil
		}
		return err
	}
	return nil
}

func (task *CommittedTaskStep) Status() (types.ContainerJSON, error) {
	if task.hasContainerInfo() {
		return task.readContainerInfo()
	}

	cli, err := client.NewClientWithOpts()
	if err != nil {
		logger.Warnf("unable to create docker client, error: %v", err)
		return types.ContainerJSON{}, err
	}

	cli.NegotiateAPIVersion(context.Background())
	containerInfo, err := cli.ContainerInspect(context.Background(), task.ContainerId)
	if err != nil {
		logger.Errorf("failed to inspect for container: %s, error: %v ",
			utils.ShortContainerId(task.ContainerId), err)
		return types.ContainerJSON{}, err
	}

	return containerInfo, nil
}

func (task *CommittedTaskStep) TaskStepDir() string {
	return GetTaskStepDir(task.EnvId, task.TaskId, task.Step)
}

func (task *CommittedTaskStep) containerInfoPath() string {
	return filepath.Join(task.TaskStepDir(), "container.json")
}

func (task *CommittedTaskStep) writeContainerInfo(info *types.ContainerJSON) error {
	task.containerInfoLock.Lock()
	defer task.containerInfoLock.Unlock()

	fp, err := os.OpenFile(task.containerInfoPath(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()
	return json.NewEncoder(fp).Encode(info)
}

func (task *CommittedTaskStep) hasContainerInfo() bool {
	task.containerInfoLock.RLock()
	defer task.containerInfoLock.RUnlock()

	_, err := os.Stat(task.containerInfoPath())
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		panic(err)
	}
	return true
}

func (task *CommittedTaskStep) readContainerInfo() (info types.ContainerJSON, err error) {
	task.containerInfoLock.RLock()
	defer task.containerInfoLock.RUnlock()

	fp, err := os.Open(task.containerInfoPath())
	if err != nil {
		return info, err
	}
	defer fp.Close()

	err = json.NewDecoder(fp).Decode(&info)
	return info, err
}

// Wait 等待任务结束返回退出码，若超时返回 error=context.DeadlineExceeded
// 如果等待到任务结束则会将容器状态信息写入到文件，然后删除容器
func (task *CommittedTaskStep) Wait(ctx context.Context) (int64, error) {
	logger := logger.WithField("taskId", task.TaskId).
		WithField("containerId", utils.ShortContainerId(task.ContainerId))

	if task.hasContainerInfo() {
		info, err := task.readContainerInfo()
		if err != nil {
			return 0, err
		}
		return int64(info.State.ExitCode), nil
	}

	cli, err := client.NewClientWithOpts()
	if err != nil {
		return 0, err
	}

	cli.NegotiateAPIVersion(ctx)
	respCh, errCh := cli.ContainerWait(ctx, task.ContainerId, container.WaitConditionNotRunning)
	select {
	case resp := <-respCh:
		if resp.Error != nil {
			logger.Warnf("wait container response status: %v, error: %v", resp.StatusCode, resp.Error)
			return resp.StatusCode, fmt.Errorf(resp.Error.Message)
		} else {
			{ // 执行结束后的处理
				// 调用 Status() 获取一次任务最新状态，并保存状态到文件
				if info, err := task.Status(); err != nil {
					logger.Warnf("get task status error: %v", err)
				} else if err := task.writeContainerInfo(&info); err != nil {
					logger.Warnf("write container info error: %v", err)
				}

				autoRemove := utils.GetBoolEnv("IAC_AUTO_REMOVE", true)
				if autoRemove {
					// 删除容器
					err := cli.ContainerRemove(context.Background(), task.ContainerId,
						types.ContainerRemoveOptions{
							RemoveVolumes: true,
							RemoveLinks:   false,
							Force:         false,
						})
					if err != nil {
						logger.Warnf("remove container error: %v", err)
					}
				}
			}

			return resp.StatusCode, nil
		}
	case err := <-errCh:
		if errdefs.IsNotFound(err) {
			logger.Infof("container not found, Id: %s", task.ContainerId)
			return 0, nil
		}
		logger.Warnf("wait container error: %v", err)
		return 0, err
	}
}
