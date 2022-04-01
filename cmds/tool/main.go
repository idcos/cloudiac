// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package main

import (
	"cloudiac/cmds/common"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

type Option struct {
	Config string `short:"c" long:"config"  default:"config-portal.yml" description:"portal config file"`
	//Verbose        []bool         `short:"v" long:"verbose" description:"Show verbose debug message"`

	ChangePassword  ChangePassword        `command:"password" description:"update user password"`
	Version         common.VersionCommand `command:"version" description:"show version"`
	InitDemo        InitDemo              `command:"init-demo" description:"init demo data with config file"`
	Scan            ScanCmd               `command:"scan" description:"scan template with policy"`
	Parse           ParseCmd              `command:"parse" description:"parse rego"`
	Upgrade2v0dot10 Update2v0dot10Cmd     `command:"upgrade2v0.10" description:"update data to v0.10"`
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
