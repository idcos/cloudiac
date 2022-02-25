// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package runner

import (
	"fmt"
	"sync"

	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

var (
	defaultDockerClient         *client.Client
	defaultDockerClientInitOnce sync.Once
)

func DockerClient() (*client.Client, error) {
	return dockerClient()
}

func dockerClient() (*client.Client, error) {
	defaultDockerClientInitOnce.Do(func() {
		var err error
		defaultDockerClient, err = initDockerClient()
		if err != nil {
			panic(fmt.Errorf("init docker client error: %v", err))
		}
	})
	return defaultDockerClient, nil
}

func initDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "create docker client")
	}
	return cli, nil
}
