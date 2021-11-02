// Copyright 2021 CloudJ Company Limited. All rights reserved.

package rbac

import (
	"cloudiac/configs"
	"cloudiac/utils/logs"
	"strings"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

var (
	enforcer *casbin.Enforcer
	initOnce sync.Once
)

// InitPolicy 初始化权限策略
func InitPolicy() {
	initOnce.Do(func() {
		logger := logs.Get().WithField("func", "initPolicy")
		logger.Infoln("init rbac policies ...")
		var (
			cModel model.Model
			err    error
		)

		// 加载策略模型
		cModel, err = model.NewModelFromString(configs.RbacModel)
		if err != nil {
			logger.Panicf("casbin load model: %v", err)
		}

		enforcer, err = casbin.NewEnforcer(cModel)
		if err != nil {
			logger.Panicf("create enforcer: %v", err)
		}

		// 加载策略
		for _, policy := range configs.Polices {
			for _, act := range strings.Split(policy.Act, "/") {
				logger.Tracef("add policy: %s %s %s", policy.Sub, policy.Obj, act)
				if _, err := enforcer.AddPolicy(policy.Sub, policy.Obj, act); err != nil {
					logger.Panicf("add policy: %v", err)
				}
			}
		}
	})
}

func Enforce(vals ...interface{}) (bool, error) {
	InitPolicy()
	return enforcer.Enforce(vals...)
}
