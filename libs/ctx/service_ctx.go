package ctx

import (
	"fmt"
	"math/rand"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/utils/logs"
)

type ServiceCtx struct {
	RC     RequestContextInter
	dbSess *db.Session
	//rdb    *cache.Session
	logger logs.Logger

	Token      string
	OrgId      uint
	UserId     uint // 登陆用户ID
	Username   string
	IsAdmin    bool
	Role       string
	user       *models.User
	UserIpAddr string
	UserAgent  string
	Perms      []string
}

func NewServiceCtx(rc RequestContextInter) *ServiceCtx {
	sc := &ServiceCtx{
		RC:     rc,
		dbSess: nil,
	}

	// 使用一个六位随机数字做为 request id
	logger := logs.Get().WithField("req", fmt.Sprintf("%08d", rand.Intn(100000000)))

	sc.logger = logger
	rc.BindServiceCtx(sc)
	return sc
}

func (c *ServiceCtx) DB() *db.Session {
	if c.dbSess == nil {
		c.dbSess = db.Get()
	}
	return c.dbSess
}

func (c *ServiceCtx) OrgDB() *db.Session {
	return c.DB().Where("org_id = ?", c.OrgId)
}

func (c *ServiceCtx) Tx() *db.Session {
	return c.DB().Begin()
}

func (c *ServiceCtx) OrgTx() *db.Session {
	return c.Tx().Where("org_id = ?", c.OrgId)
}

func (c *ServiceCtx) Logger() logs.Logger {
	return c.logger
}

func (c *ServiceCtx) MustUser() *models.User {
	user, err := c.User()
	if err != nil {
		panic(err)
	}
	return user
}

func (c *ServiceCtx) User() (*models.User, error) {
	if c.user == nil {
		var user models.User
		err := c.DB().Where("id = ?", c.UserId).First(&user)
		if err != nil {
			return nil, err
		}
		c.user = &user
	}
	return c.user, nil
}

func (c *ServiceCtx) AddLogField(key string, val string) *ServiceCtx {
	c.logger = c.logger.WithField(key, val)
	return c
}

//func (c *ServiceCtx) Cache() *cache.Session {
//	if c.rdb == nil {
//		c.rdb = cache.Client()
//	}
//
//	return c.rdb
//}
