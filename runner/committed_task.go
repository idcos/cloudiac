// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package runner

import (
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
)

var logger = logs.Get()

type StartedTask struct {
	EnvId       string `json:"envId"`
	TaskId      string `json:"taskId"`
	Step        int    `json:"step"`
	ContainerId string `json:"containerId"`
	ExecId      string `json:"execId"`

	StartedAt *time.Time `json:"startedAt"`
	Timeout   int        `json:"timeout"`

	PauseOnFinish bool `json:"pauseOnFinish"` // 该步骤结束时暂停容器

	containerInfoLock sync.RWMutex `json:"-"`
}

func LoadStartedTask(envId string, taskId string, step int) (*StartedTask, error) {
	path := filepath.Join(GetTaskDir(envId, taskId, step), TaskInfoFileName)
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	task := StartedTask{}
	if err := json.NewDecoder(fp).Decode(&task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (task *StartedTask) Cancel() error {
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

func (task *StartedTask) Status() (info types.ContainerExecInspect, err error) {
	if task.hasContainerInfo() {
		return task.readContainerInfo()
	}

	info, err = Executor{}.GetExecInfo(task.ExecId)
	if err != nil {
		return info, err
	}
	return info, nil
}

func (task *StartedTask) TaskDir() string {
	return GetTaskDir(task.EnvId, task.TaskId, task.Step)
}

func (task *StartedTask) containerInfoPath() string {
	return filepath.Join(task.TaskDir(), TaskContainerInfoFileName)
}

func (task *StartedTask) writeContainerInfo(info *types.ContainerExecInspect) error {
	task.containerInfoLock.Lock()
	defer task.containerInfoLock.Unlock()

	fp, err := os.OpenFile(task.containerInfoPath(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()
	return json.NewEncoder(fp).Encode(info)
}

func (task *StartedTask) hasContainerInfo() bool {
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

func (task *StartedTask) readContainerInfo() (info types.ContainerExecInspect, err error) {
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
// 如果等待到任务结束则会将容器状态信息写入到文件，并判断是否需要暂停容器
// 注意：该函数可能会被多个请求源同时调用，不要在该函数中添加不可重复执行的逻辑。
func (task *StartedTask) Wait(ctx context.Context) (int64, error) {
	logger := logger.WithField("taskId", task.TaskId).
		WithField("containerId", utils.ShortContainerId(task.ContainerId))

	if task.hasContainerInfo() {
		info, err := task.readContainerInfo()
		if err != nil {
			return 0, err
		}
		return int64(info.ExitCode), nil
	}

	var (
		info types.ContainerExecInspect
		err  error
	)
	if task.StartedAt != nil && task.Timeout > 0 {
		deadline := task.StartedAt.Add(time.Duration(task.Timeout) * time.Second)
		info, err = Executor{}.WaitCommandWithDeadline(ctx, task.ContainerId, task.ExecId, deadline)
	} else {
		info, err = Executor{}.WaitCommand(ctx, task.ContainerId, task.ExecId)
	}

	if err != nil {
		logger.Warnf("wait step error: %v", err)
		return 0, err
	}

	// 执行结束后的处理
	{
		// 调用 Status() 获取一次任务最新状态，并保存状态到文件
		if info, err = task.Status(); err != nil {
			logger.Warnf("get task status error: %v", err)
		} else if err := task.writeContainerInfo(&info); err != nil {
			logger.Warnf("write container info error: %v", err)
		}

		// 暂停容器
		// 当有多个请求源同时调用该函数时，容器暂停操作可能被调用多次，这是允许的，多次调用 Pause() 不会报错。
		// 但要保证调用 Pause() 是及时的，避免下一步骤己经启动了，前一步骤触发的 Pause 操作才被调用，这会导致容器被异常暂停。
		if task.PauseOnFinish {
			logger.Debugf("pause container %s", utils.ShortContainerId(info.ContainerID))
			if err := (Executor{}).Pause(info.ContainerID); err != nil {
				logger.Warn(err)
			}
		}
	}

	return int64(info.ExitCode), nil
}
