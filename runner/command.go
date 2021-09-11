// Copyright 2021 CloudJ Company Limited. All rights reserved.

package runner

import (
	"context"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"

	"cloudiac/common"
	"cloudiac/configs"
	"cloudiac/utils"
)

// Command to run docker image
type Command struct {
	Image      string
	Env        []string
	Timeout    int
	PrivateKey string

	TerraformVersion string
	Commands         []string
	HostWorkdir      string // 宿主机目录
	Workdir          string // 容器目录
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

	// 内置 tf 版本列表中无该版本，我们挂载缓存目录到容器，下载后会保存到宿主机，下次可以直接使用。
	// 注意，该方案有个问题：客户无法自定义镜像预先安装需要的 terraform 版本，
	// 因为判断版本不在 TerraformVersions 列表中就会挂载目录，客户自定义镜像安装的版本会被覆盖
	//（考虑把版本列表写到配置文件？）
	if !utils.StrInArray(cmd.TerraformVersion, common.TerraformVersions...) {
		mountConfigs = append(mountConfigs, mount.Mount{
			Type:   mount.TypeBind,
			Source: conf.Runner.AbsTfenvVersionsCachePath(),
			Target: "/root/.tfenv/versions",
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
