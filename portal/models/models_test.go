// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/portal/libs/db"
	"fmt"
	"testing"
)

func TestCreateBatch(t *testing.T) {
	mysqlConfig := "root:123456@tcp(127.0.0.1:3306)/iac?charset=utf8mb4&parseTime=True&loc=Local"
	db.Init(mysqlConfig)
	m := []Modeler{
		&Project{Name: "1"},
		&Project{Name: "12"},
		&Project{Name: "13"},
		&Project{Name: "14"},
		&Project{Name: "15"},
		&Project{Name: "16"},
		&Project{Name: "17"},
		&Project{Name: "18"},
	}

	err := CreateBatch(db.Get().Table("iac_project"), m)

	fmt.Println(err)
}
