// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

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
	Scan            ScanCmd               `command:"scan" description:"scan template with policy"`
	Parse           ParseCmd              `command:"parse" description:"parse rego"`
	Upgrade2v0dot10 Update2v0dot10Cmd     `command:"upgrade2v0.10" description:"update data to v0.10"`
	Bill            BillCmd               `command:"bill-collect" description:"bill collect"`
	DumpDb          DumpDb                `command:"dumpdb" description:"dump db to yaml"`
	InitDB          InitDB                `command:"initdb" description:"init database structure"`
	UpdateDb        UpdateDb              `command:"updateDB" description:"update database data"`
	Yaml2Hcl        Yaml2HclCmd           `command:"yaml2hcl" description:"yaml to hcl"`

	// 初始化演示项目。
	// 旧版本中通过这个命令来创建一个共用的演示项目，但在 0.12 版本演示项目改为了为每个用户单独创建，所以废弃该命令
	// InitDemo        InitDemo              `command:"init-demo" description:"init demo data with config file"`
}

var (
	// iac-tool 的 logger 单独设置
	logger = logrus.New()
	opt    = Option{}
)

func main() {
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000",
	})
	common.LoadDotEnv()

	_, err := flags.Parse(&opt)
	if err != nil {
		os.Exit(1)
	}
}
