// Copyright 2021 CloudJ Company Limited. All rights reserved.

package main

import (
	"cloudiac/configs"
	"cloudiac/portal/libs/db"
	"cloudiac/utils/logs"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"strings"
)

// initPolicy 初始化权限策略
func initPolicy(tx *db.Session) error {
	logger := logs.Get().WithField("func", "initPolicy")
	logger.Infoln("init rbac policy...")
	var err error

	adapter, err := gormadapter.NewAdapterByDBUseTableName(tx.GormDB(), "iac_", "")
	if err != nil {
		panic(fmt.Sprintf("error create enforcer: %v", err))
	}

	// 加载策略模型
	m, err := model.NewModelFromString(configs.RbacModel)
	if err != nil {
		panic(fmt.Sprintf("error load rbac model: %v", err))
	}
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		panic(fmt.Sprintf("error create enforcer: %v", err))
	}

	// 初始化策略数据库
	for _, policy := range configs.Polices {
		for _, act := range strings.Split(policy.Act, "/") {
			logger.Debugf("add policy: %s %s %s", policy.Sub, policy.Obj, act)
			enforcer.AddPolicy(policy.Sub, policy.Obj, act)
		}
	}

	return nil
}
