// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package ctx

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils/logs"
	"crypto/rand"
	"fmt"
	"math/big"
)

type ServiceContext struct {
	rc     RequestContext
	dbSess *db.Session
	logger logs.Logger

	UserId       models.Id // 登陆用户ID
	OrgId        models.Id // 组织ID
	ProjectId    models.Id // 项目ID
	Email        string    // 用户邮箱
	Username     string    // 用户名称
	IsSuperAdmin bool      // 是否平台管理员
	UserIpAddr   string
}

func NewServiceContext(rc RequestContext) *ServiceContext {
	bigNum, err := rand.Int(rand.Reader, big.NewInt(100000000))
	if err != nil {
		panic(fmt.Errorf("get random number err: %+v", err))
	}
	logger := logs.Get().WithField("req", fmt.Sprintf("%08d", bigNum))

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
