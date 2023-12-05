// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.
package main

import (
	"cloudiac/configs"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"

	"github.com/go-testfixtures/testfixtures/v3"
)

// ./iac-tool DumpDb ./testdata

type DumpDb struct{}

func (*DumpDb) Execute(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("output directory is required")
	}

	outputDir := args[0]
	configs.Init(opt.Config)
	fmt.Printf("args %+v\n", configs.Get().Dsn())
	db.Init(configs.Get().Dsn())
	models.Init(true)
	sqlDb, _ := db.Get().GormDB().DB()

	// 只初始化数据库不做导出
	if outputDir == "init-only" {
		return nil
	}

	dumper, err := testfixtures.NewDumper(
		testfixtures.DumpDatabase(sqlDb),
		testfixtures.DumpDialect("mysql"), // or your database of choice
		testfixtures.DumpDirectory(outputDir),
		// testfixtures.DumpTables( // optional, will dump all table if not given
		// 	"posts",
		// 	"comments",
		// 	"tags",
		// ),
	)
	if err != nil {
		panic(err)
	}
	if err := dumper.Dump(); err != nil {
		panic(err)
	}

	return nil
}
