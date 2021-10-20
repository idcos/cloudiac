// Copyright 2021 CloudJ Company Limited. All rights reserved.

package ctx

import (
	"cloudiac/configs"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils/logs"
	"fmt"
	"math/rand"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

type ServiceContext struct {
	rc       RequestContext
	dbSess   *db.Session
	logger   logs.Logger
	enforcer *casbin.Enforcer

	UserId       models.Id // 登陆用户ID
	OrgId        models.Id // 组织ID
	ProjectId    models.Id // 项目ID
	Username     string    // 用户名称
	IsSuperAdmin bool      // 是否平台管理员
	UserIpAddr   string
}

func NewServiceContext(rc RequestContext) *ServiceContext {
	logger := logs.Get().WithField("req", fmt.Sprintf("%08d", rand.Intn(100000000)))

	sc := &ServiceContext{
		rc:     rc,
		dbSess: nil,
		logger: logger,
	}

	rc.BindService(sc)
	return sc
}

func (c *ServiceContext) DB() *db.Session {
	if c.dbSess == nil {
		c.dbSess = db.Get()
	}
	return c.dbSess
}

func (c *ServiceContext) Tx() *db.Session {
	return c.DB().Begin()
}

func (c *ServiceContext) Logger() logs.Logger {
	return c.logger
}

func (c *ServiceContext) AddLogField(key string, val string) *ServiceContext {
	c.logger = c.logger.WithField(key, val)
	return c
}

// Enforcer 初始化 casbin enforcer 对象
func (c *ServiceContext) Enforcer() *casbin.Enforcer {
	if c.enforcer == nil {
		var err error

		adapter, err := gormadapter.NewAdapterByDBUseTableName(db.Get().GormDB(), "iac_", "")
		if err != nil {
			panic(fmt.Sprintf("error create enforcer: %v", err))
		}

		// 加载策略模型
		m, err := model.NewModelFromString(configs.RbacModel)
		if err != nil {
			panic(fmt.Sprintf("error load rbac model: %v", err))
		}
		c.enforcer, err = casbin.NewEnforcer(m, adapter)
		if err != nil {
			panic(fmt.Sprintf("error create enforcer: %v", err))
		}
	}
	return c.enforcer
}
