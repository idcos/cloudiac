package runner

import (
	"cloudiac/configs"
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	guuid "github.com/google/uuid"
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

func (cmd *Command) Create(dirMapping string) error {
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
					Source: conf.Runner.AssetPath,
					Target: "/assets",
				},
				{
					Type:   mount.TypeBind,
					Source: dirMapping,
					Target: ContainerLogFilePath,
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
