// Copyright 2021 CloudJ Company Limited. All rights reserved.

package main

import (
	"cloudiac/cmds/common"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	"os"
)

type Option struct {
	Config string `short:"c" long:"config"  default:"config-portal.yml" description:"portal config file"`
	//Verbose        []bool         `short:"v" long:"verbose" description:"Show verbose debug message"`

	ChangePassword ChangePassword        `command:"password" description:"update user password"`
	Version        common.VersionCommand `command:"version" description:"show version"`
	InitDemo       InitDemo              `command:"init-demo" description:"init demo data with config file"`
	Policy         PolicyCmd             `command:"policy" description:"check template with policy"`
}

var (
	// iac-tool 的 logger 单独设置
	logger = logrus.New()
	opt    = Option{}
)

func main() {
	common.LoadDotEnv()

	_, err := flags.Parse(&opt)
	if err != nil {
		os.Exit(1)
	}
}
