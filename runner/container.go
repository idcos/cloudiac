package runner

import (
	"cloudiac/configs"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	guuid "github.com/google/uuid"
	"log"
	"os"
	"path/filepath"
)

func (task *CommitedTask) Cancel() error {
	cli, err := client.NewClientWithOpts()
	cli.NegotiateAPIVersion(context.Background())
	if err != nil {
		log.Printf("Unable to create docker client")
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

func (task *CommitedTask) Status() (types.ContainerJSON, error) {
	if task.hasContainerInfo() {
		return task.readContainerInfo()
	}

	cli, err := client.NewClientWithOpts()
	if err != nil {
		log.Printf("Unable to create docker client, error: %v", err)
		return types.ContainerJSON{}, err
	}

	cli.NegotiateAPIVersion(context.Background())
	containerInfo, err := cli.ContainerInspect(context.Background(), task.ContainerId)
	if err != nil {
		log.Printf("Failed to inspect for container id: %s, error: %v ", task.ContainerId, err)
		return types.ContainerJSON{}, err
	}

	return containerInfo, nil
}

func (task *CommitedTask) workdir() string {
	return GetTaskWorkDir(task.TemplateId, task.TaskId)
}

func (task *CommitedTask) containerInfoPath() string {
	return filepath.Join(task.workdir(), "container.json")
}

func (task *CommitedTask) writeContainerInfo(info *types.ContainerJSON) error {
	task.containerInfoLock.Lock()
	defer task.containerInfoLock.Unlock()

	fp, err := os.OpenFile(task.containerInfoPath(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()
	return json.NewEncoder(fp).Encode(info)
}

func (task *CommitedTask) hasContainerInfo() bool {
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

func (task *CommitedTask) readContainerInfo() (info types.ContainerJSON, err error) {
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
func (task *CommitedTask) Wait(ctx context.Context) (int64, error) {
	logger := logs.Get().WithField("taskId", task.TaskId).
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
			logger.Debugf("container not found, Id: %s", task.ContainerId)
			return 0, nil
		}
		logger.Warnf("wait container error: %v", err)
		return 0, err
	}
}

func (cmd *Command) Create() error {
	// TODO(ZhengYue): Create client with params of host info
	cli, err := client.NewClientWithOpts()
	cli.NegotiateAPIVersion(context.Background())

	if err != nil {
		log.Printf("Unable to create docker client")
		return err
	}

	id := guuid.New()
	conf := configs.Get()

	log.Printf("starting command, task working directory: %s", cmd.TaskWorkdir)
	cont, err := cli.ContainerCreate(
		cmd.ContainerInstance.Context,
		&container.Config{
			Image:        cmd.Image,
			Cmd:          cmd.Commands,
			Env:          cmd.Env,
			AttachStdout: true,
			AttachStderr: true,
		},
		&container.HostConfig{
			AutoRemove: false,
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: cmd.TaskWorkdir,
					Target: ContainerTaskDir,
				},
				{
					Type:     mount.TypeBind,
					Source:   conf.Runner.AbsAssetsPath(),
					Target:   ContainerAssetsDir,
					ReadOnly: true,
				},
				{
					// providers 需要挂载到指定目录才能被 terraform 查找到，所以单独再做一次挂载
					Type:     mount.TypeBind,
					Source:   conf.Runner.ProviderPath(),
					Target:   ContainerPluginsPath,
					ReadOnly: true,
				},
				{
					Type:   mount.TypeBind,
					Source: conf.Runner.AbsPluginCachePath(),
					Target: ContainerPluginsCachePath,
				},
				{
					Type:   mount.TypeBind,
					Source: "/var/run/docker.sock",
					Target: "/var/run/docker.sock",
				},
			},
		},
		nil,
		nil,
		id.String())
	if err != nil {
		log.Printf("ContainerCreate err: %v", err)
		return err
	}

	cmd.ContainerInstance.ID = cont.ID
	log.Printf("Create container ID = %s", cont.ID)
	err = cli.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	return err
}
