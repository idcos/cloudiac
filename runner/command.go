// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pkg/errors"

	"cloudiac/common"
	"cloudiac/configs"
	"cloudiac/utils"
)

// Task Executor
type Executor struct {
	Image      string
	Name       string
	Env        []string
	Timeout    int
	PrivateKey string

	TerraformVersion string
	Commands         []string
	HostWorkdir      string // 宿主机目录
	Workdir          string // 容器目录
	AutoRemove       bool   // 开启容器的自动删除？
	// for container
	//ContainerInstance *Container
}

// Container Info
type Container struct {
	Context context.Context
	ID      string
	RunID   string
}

func (exec *Executor) tryPullImage(cli *client.Client) {
	logger := logger.WithField("image", exec.Image).WithField("action", "TryPullImage")
	if cli == nil {
		var err error
		cli, err = dockerClient()
		if err != nil {
			logger.Warn(err)
			return
		}
	}

	reader, err := cli.ImagePull(context.Background(), exec.Image, types.ImagePullOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Debugf("pull image: %v", err)
		} else {
			logger.Warnf("pull image: %v", err)
		}
		return
	}
	defer reader.Close()

	bs, _ := ioutil.ReadAll(reader)
	logger.Tracef("pull image: %s", bs)
}

func (exec *Executor) Start() (string, error) {
	logger := logger.WithField("taskId", filepath.Base(exec.HostWorkdir))
	cli, err := dockerClient()
	if err != nil {
		logger.Error(err)
		return "", err
	}
	logger.Infof("pull image: %s", exec.Image)
	// TODO: 补充 pull 失败的错误处理
	exec.tryPullImage(cli)

	conf := configs.Get()
	mountConfigs := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: exec.HostWorkdir,
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

	if conf.Consul.ConsulTls {
		mountConfigs = append(mountConfigs, mount.Mount{
			Type:   mount.TypeBind,
			Source: conf.Consul.ConsulCertPath,
			Target: ContainerCertificateDir,
		})
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
	if !utils.StrInArray(exec.TerraformVersion, common.TerraformVersions...) {
		mountConfigs = append(mountConfigs, mount.Mount{
			Type:   mount.TypeBind,
			Source: conf.Runner.AbsTfenvVersionsCachePath(),
			Target: "/root/.tfenv/versions",
		})
	}

	c, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:        exec.Image,
			WorkingDir:   exec.Workdir,
			Cmd:          exec.Commands,
			Env:          exec.Env,
			OpenStdin:    true,
			Tty:          true,
			AttachStdin:  false,
			AttachStdout: true,
			AttachStderr: true,
		},
		&container.HostConfig{
			AutoRemove: exec.AutoRemove,
			Mounts:     mountConfigs,
		},
		nil,
		nil,
		exec.Name)
	if err != nil {
		logger.Errorf("create container err: %v", err)
		return "", err
	}

	cid := utils.ShortContainerId(c.ID)
	logger.Infof("container id: %s", cid)
	err = cli.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{})
	return cid, err
}

func (Executor) RunCommand(cid string, command []string) (execId string, err error) {
	cli, err := dockerClient()
	if err != nil {
		return "", err
	}

	resp, err := cli.ContainerExecCreate(context.Background(), cid, types.ExecConfig{
		Detach: false,
		Cmd:    command,
	})
	if err != nil {
		err = errors.Wrap(err, "container exec create")
		return "", err
	}

	err = cli.ContainerExecStart(context.Background(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		err = errors.Wrap(err, "container exec start")
		return "", err
	}

	return resp.ID, nil
}

// 执行命令并获取输出
func (Executor) RunCommandOutput(cid string, command []string) (output []byte, err error) {
	cli, err := dockerClient()
	if err != nil {
		return nil, err
	}

	resp, err := cli.ContainerExecCreate(context.Background(), cid, types.ExecConfig{
		AttachStdin:  false,
		AttachStderr: true,
		AttachStdout: true,
		Detach:       false,
		Cmd:          command,
	})
	if err != nil {
		err = errors.Wrap(err, "container exec create")
		return nil, err
	}

	hijackedResp, err := cli.ContainerExecAttach(context.Background(), resp.ID, types.ExecStartCheck{})
	if err != nil {
		err = errors.Wrap(err, "container exec start")
		return nil, err
	}
	defer hijackedResp.Close()

	buffer := bytes.NewBuffer(nil)
	_, err = stdcopy.StdCopy(buffer, buffer, hijackedResp.Reader)
	if err != nil && !errors.Is(err, io.EOF) {
		return buffer.Bytes(), err
	}
	return buffer.Bytes(), nil
}

func (Executor) GetExecInfo(execId string) (execInfo types.ContainerExecInspect, err error) {
	cli, err := dockerClient()
	if err != nil {
		return execInfo, err
	}
	execInfo, err = cli.ContainerExecInspect(context.Background(), execId)
	if err != nil {
		return execInfo, errors.Wrap(err, "container exec attach")
	}
	return execInfo, nil
}

func (Executor) Wait(ctx context.Context, cid string) error {
	cli, err := dockerClient()
	if err != nil {
		return err
	}

	okCh, errCh := cli.ContainerWait(ctx, cid, container.WaitConditionNotRunning)
	select {
	case <-okCh:
		return nil
	case err = <-errCh:
		return err
	}
}

var ErrContainerNotRun = fmt.Errorf("container not running")
var ErrTaskAborted = fmt.Errorf("task aborted")

func (Executor) WaitCommand(ctx context.Context, containerId string, execId string) (execInfo types.ContainerExecInspect, err error) {
	cli, err := dockerClient()
	if err != nil {
		return execInfo, err
	}

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return execInfo, ctx.Err()
		case <-ticker.C:
		default:
		}

		if ci, err := cli.ContainerInspect(ctx, containerId); err != nil {
			return execInfo, errors.Wrap(err, "container inspect")
		} else if ci.State.Paused || !ci.State.Running {
			return execInfo, errors.Wrapf(ErrContainerNotRun, "container status is %s", ci.State.Status)
		}

		inspect, err := cli.ContainerExecInspect(ctx, execId)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return execInfo, err
			}
			return execInfo, errors.Wrap(err, "container exec inspect")
		}
		if !inspect.Running {
			return execInfo, nil
		}
	}
}

// 等待进程结束，如果提前触发了 deadline 则 kill 进程
func (exec Executor) WaitCommandWithDeadline(ctx context.Context, containerId string, execId string, deadline time.Time) (execInfo types.ContainerExecInspect, err error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, deadline)
	defer cancel()

	logger.Debugf("wait exec %s, deadline: %s", execId, deadline.Format(time.RFC3339))
	if execInfo, err = exec.WaitCommand(ctx, containerId, execId); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			// logger.Infof("task %s/step%s: %v", exec..TaskId, t.req.Step, err)
			if err := (Executor{}).StopCommand(execId); err != nil {
				logger.WithField("cid", execInfo.ContainerID).Errorf("stop command error: %v", err)
			}
		}
		return execInfo, err
	}

	return execInfo, err
}

func (e Executor) StopCommand(execId string) (err error) {
	cli, err := dockerClient()
	if err != nil {
		return err
	}

	inspect, err := cli.ContainerExecInspect(context.Background(), execId)
	if err != nil {
		return errors.Wrap(err, "container exec attach")
	}

	// 先执行 kill，等待 30s，然后 kill -9
	if _, err := e.RunCommand(inspect.ContainerID, []string{
		"sh", "-c",
		fmt.Sprintf(
			"for i in `seq 1 30`;do kill %d && sleep 1 || break; done; kill -9 %d",
			inspect.Pid, inspect.Pid),
	}); err != nil {
		return errors.Wrap(err, "kill process")
	}
	return nil
}

func (Executor) IsPaused(cid string) (bool, error) {
	cli, err := dockerClient()
	if err != nil {
		return false, err
	}

	inspect, err := cli.ContainerInspect(context.Background(), cid)
	if err != nil {
		return false, errors.Wrapf(err, "%s, container inspect", cid)
	}

	return inspect.State.Paused, nil
}

func (Executor) Pause(cid string) (err error) {
	cli, err := dockerClient()
	if err != nil {
		return err
	}

	if err := cli.ContainerPause(context.Background(), cid); err != nil {
		if strings.Contains(err.Error(), "is not running") ||
			strings.Contains(err.Error(), "is already paused") {
			return nil
		}
		err = errors.Wrapf(err, "pause container %s", cid)
		return err
	}

	return nil
}

func (Executor) Unpause(cid string) (err error) {
	cli, err := dockerClient()
	if err != nil {
		return err
	}

	if err := cli.ContainerUnpause(context.Background(), cid); err != nil {
		err = errors.Wrapf(err, "unpause container %s", cid)
		return err
	}
	return nil
}

func (Executor) UnpauseIf(cid string) (err error) {
	if ok, err := (Executor{}).IsPaused(cid); err != nil {
		return err
	} else if ok {
		logger.Debugf("unpause container %s", cid)
		if err := (Executor{}).Unpause(cid); err != nil {
			return err
		}
		logger.Debugf("unpause container %s, done", cid)
	}
	return nil
}
