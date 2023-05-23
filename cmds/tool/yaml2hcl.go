// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package main

import (
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"github.com/spf13/viper"
)

type Yaml2HclCmd struct {
	Config string `long:"config" short:"c" description:"yaml file" required:"true"`
}

func (*Yaml2HclCmd) Usage() string {
	return `<yaml to hcl>`
}

func (y *Yaml2HclCmd) Execute(args []string) error {
	var configViperConfig = viper.New()
	configViperConfig.SetConfigFile(y.Config)
	configViperConfig.SetConfigType("yaml")
	configViperConfig.SetConfigType("yml")
	//读取配置文件内容
	if err := configViperConfig.ReadInConfig(); err != nil {
		return err
	}
	var deploy forms.DeployForm
	if err := configViperConfig.Unmarshal(&deploy); err != nil {
		return err
	}

	hclStr := services.Yaml2Hcl(deploy)
	fmt.Println(hclStr)

	return nil
}
