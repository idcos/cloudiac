// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package main

import (
	v1 "cloudiac/runner/api/v1"
	"cloudiac/utils"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"

	"cloudiac/cmds/common"
	iac_common "cloudiac/common"
	"cloudiac/configs"
	"cloudiac/utils/logs"
)

type Option struct {
	common.OptionVersion

	Config     string `short:"c" long:"config"  default:"config-runner.yml" description:"config file"`
	Verbose    []bool `short:"v" long:"verbose" description:"Show verbose debug message"`
	ReRegister bool   `long:"re-register" description:"Re registration service to Consul"`
}

func main() {
	common.LoadDotEnv()

	opt := Option{}
	_, err := flags.Parse(&opt)
	if err != nil {
		os.Exit(1)
	}
	common.ShowVersionIf(opt.Version)

	configs.Init(opt.Config, configs.ParseRunnerConfig)
	if err := checkConfigs(configs.Get()); err != nil {
		panic(err)
	}
	if err := ensureDirs(); err != nil {
		panic(err)
	}

	logConf := configs.Get().Log
	logs.Init(logConf.LogLevel, logConf.LogPath, logConf.LogMaxDays)

	runnerConfJson, _ := json.Marshal(configs.Get().Runner)
	logs.Get().Infof("runner configs: %s", runnerConfJson)

	common.ReRegisterService(opt.ReRegister, iac_common.RunnerServiceName)
	StartServer()
}

func checkConfigs(c *configs.Config) error {
	cases := []struct {
		name  string
		value string
	}{
		{"runner.default_image", c.Runner.DefaultImage},
		{"runner.storage_path", c.Runner.StoragePath},
		{"runner.plugin_cache_path", c.Runner.PluginCachePath},
	}

	for _, c := range cases {
		if c.value == "" {
			return fmt.Errorf("configuration '%s' is empty", c.name)
		}
	}
	return nil
}

// ensureDirs 确保依赖的目录存在
func ensureDirs() error {
	c := configs.Get().Runner

	var err error
	for _, path := range []string{
		c.StoragePath,
		c.AssetsPath,
		c.PluginCachePath,
		c.ProviderPath(),
		c.AbsTfenvVersionsCachePath(),
	} {
		if path == "" {
			continue
		}

		// 确保可以转为绝对路径，因为挂载到容器中时必须使用绝对路径
		path, err = filepath.Abs(path)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Abs(%s)", path))
		} else if err = os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

func StartServer() {
	conf := configs.Get()
	logger := logs.Get()

	utils.SetGinMode()
	e := gin.New()
	e.Use(gin.RecoveryWithWriter(io.MultiWriter(
		gin.DefaultWriter,
		logs.MustGetLogWriter("error"),
	)))

	v1.RegisterRoute(e.Group("/api/v1"))
	logger.Infof("starting runner on %v", conf.Listen)
	if err := e.Run(conf.Listen); err != nil {
		logger.Fatalln(err)
	}
}
