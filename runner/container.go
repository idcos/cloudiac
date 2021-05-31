package runner

import (
	"cloudiac/configs"
	"cloudiac/utils/logs"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	guuid "github.com/google/uuid"
	"log"
)

func (task *CommitedTask) Cancel() error {
	cli, err := client.NewClientWithOpts()
	cli.NegotiateAPIVersion(context.Background())
	if err != nil {
		log.Printf("Unable to create docker client")
		panic(err)
	}

	conatinerRemoveOpts := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	if err := cli.ContainerRemove(context.Background(), task.ContainerId, conatinerRemoveOpts); err != nil {
		log.Printf("Unable to remove container: %s", err)
		return err
	}
	return nil
}

func (task *CommitedTask) Status() (types.ContainerJSON, error) {
	cli, err := client.NewClientWithOpts()
	cli.NegotiateAPIVersion(context.Background())
	if err != nil {
		log.Printf("Unable to create docker client")
		panic(err)
	}

	containerInfo, err := cli.ContainerInspect(context.Background(), task.ContainerId)
	if err != nil {
		log.Printf("Failed to inspect for container id: %s ", task.ContainerId)
		panic(err)
	}
	return containerInfo, nil
}

// Wait 等待任务结束，返回退出码，如果超时，返回 error context.DeadlineExceeded
func (task *CommitedTask) Wait(ctx context.Context) (int64, error) {
	logger := logs.Get().WithField("taskId", task.TaskId)

	cli, err := client.NewClientWithOpts()
	if err != nil {
		return 0, err
	}

	cli.NegotiateAPIVersion(ctx)
	respCh, errCh := cli.ContainerWait(ctx, task.ContainerId, container.WaitConditionNotRunning)
	select {
	case resp := <-respCh:
		logger.Debugf("wait container resp: %#v", resp)
		if resp.Error != nil {
			return resp.StatusCode, fmt.Errorf(resp.Error.Message)
		} else {
			return resp.StatusCode, nil
		}
	case err := <-errCh:
		if errdefs.IsNotFound(err) {
			logger.Debugf("container not found, Id: %s", task.ContainerId)
			return 0, nil
		}
		logger.Warnf("wait container error: %#v", err)
		return 0, err
	}
}

func (cmd *Command) Create() error {
	// TODO(ZhengYue): Create client with params of host info
	cli, err := client.NewClientWithOpts()
	cli.NegotiateAPIVersion(context.Background())

	if err != nil {
		log.Printf("Unable to create docker client")
		panic(err)
	}

	id := guuid.New()

	conf := configs.Get()

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
					Target: ContainerIaCDir,
				},
				{
					Type:   mount.TypeBind,
					Source: "/var/run/docker.sock",
					Target: "/var/run/docker.sock",
				},
				{
					Type:   mount.TypeBind,
					Source: conf.Runner.ProviderPath,
					Target: ContainerProviderPath,
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
