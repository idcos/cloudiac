// Copyright 2021 CloudJ Company Limited. All rights reserved.

package ctx

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils/logs"
	"fmt"
	"math/rand"
)

type ServiceContext struct {
	rc     RequestContext
	dbSess *db.Session
	logger logs.Logger

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
