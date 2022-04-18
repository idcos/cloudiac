// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.
package main

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

// ./iac-tool InitDB root:Yunjikeji@tcp(127.0.0.1:3307)/iac_test?charset=utf8mb4&parseTime=True&loc=Local

type InitDB struct{}

const (
	defaultDsn = "root:Yunjikeji@tcp(127.0.0.1:3307)/iac_test?charset=utf8mb4&parseTime=True&loc=Local"
)

func (*InitDB) Execute(args []string) error {
	// unit test database default configuration
	dsn := defaultDsn
	if len(args) > 0 {
		dsn = args[0]
	}

	db.Init(dsn)
	models.Init(true)

	return nil
}
