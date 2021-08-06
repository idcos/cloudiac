// Copyright 2021 CloudJ Company Limited. All rights reserved.

package runner

import (
	"cloudiac/configs"
	"cloudiac/utils"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"path/filepath"
)

// Command to run docker image
type Command struct {
	Image      string
	Env        []string
	Timeout    int
	PrivateKey string

	Commands    []string
	HostWorkdir string // 宿主机目录
	Workdir     string // 容器目录
	// for container
	//ContainerInstance *Container
}

// Container Info
type Container struct {
	Context context.Context
	ID      string
	RunID   string
}

func (cmd *Command) Start() (string, error) {
	logger := logger.WithField("taskId", filepath.Base(cmd.HostWorkdir))
	cli, err := client.NewClientWithOpts()
	if err != nil {
		logger.Errorf("unable to create docker client")
		return "", err
	}
	cli.NegotiateAPIVersion(context.Background())

	conf := configs.Get()
	mountConfigs := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: cmd.HostWorkdir,
			Target: ContainerWorkspace,
		},
		{
			Type:   mount.TypeBind,
			Source: conf.Runner.AbsPluginCachePath(),
			Target: ContainerPluginCachePath,
		},
		{
			Type:   mount.TypeBind,
			Source: "/var/run/docker.sock",
			Target: "/var/run/docker.sock",
		},
	}

	// assets_path 配置为空则表示直接使用 worker 容器中打包的 assets。
	// 在 runner 容器化部署时运行 runner 的宿主机(docker host)并没有 assets 目录，
	// 如果配置了 assets 路径，进行 bind mount 时会因为源目录不存在而报错。
	if conf.Runner.AssetsPath != "" {
		mountConfigs = append(mountConfigs, mount.Mount{
			Type:     mount.TypeBind,
			Source:   conf.Runner.AbsAssetsPath(),
			Target:   ContainerAssetsDir,
			ReadOnly: true,
		})
		mountConfigs = append(mountConfigs, mount.Mount{
			// providers 需要挂载到指定目录才能被 terraform 查找到，所以单独做一次挂载
			Type:     mount.TypeBind,
			Source:   conf.Runner.ProviderPath(),
			Target:   ContainerPluginPath,
			ReadOnly: true,
		})
	}

	c, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:        cmd.Image,
			WorkingDir:   cmd.Workdir,
			Cmd:          cmd.Commands,
			Env:          cmd.Env,
			AttachStdin:  false,
			AttachStdout: true,
			AttachStderr: true,
		},
		&container.HostConfig{
			AutoRemove: false,
			Mounts:     mountConfigs,
		},
		nil,
		nil,
		"")
	if err != nil {
		logger.Errorf("create container err: %v", err)
		return "", err
	}

	cid := utils.ShortContainerId(c.ID)
	logger.Infof("container id: %s", cid)
	err = cli.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{})
	return cid, err
}
