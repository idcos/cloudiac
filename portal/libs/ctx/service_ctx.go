package ctx

import (
	"cloudiac/configs"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils/logs"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v2"
	"math/rand"
)

type ServiceCtx struct {
	RC     RequestContextInter
	dbSess *db.Session
	//rdb    *cache.Session
	logger logs.Logger

	Token string
	OrgId models.Id
	org   *models.Organization
	//OrgGuid    string
	UserId       models.Id // 登陆用户ID
	Username     string    // 用户名称
	IsSuperAdmin bool      // 是否平台管理员
	Role         string    // 组织角色
	user         *models.User
	UserIpAddr   string
	UserAgent    string
	Perms        []string
	ProjectId    models.Id // 项目ID
	ProjectRole  string    // 项目角色

	// Casbin
	enforcer *casbin.Enforcer
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

func (c *ServiceCtx) MustOrg() *models.Organization {
	org, err := c.Org()
	if err != nil {
		panic(err)
	}
	return org
}

func (c *ServiceCtx) Org() (*models.Organization, error) {
	if c.org == nil {
		var org models.Organization
		err := c.DB().Where("id = ?", c.OrgId).First(&org)
		if err != nil {
			return nil, err
		}
		c.org = &org
	}
	return c.org, nil
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

// Enforcer 初始化 casbin enforcer 对象
func (c *ServiceCtx) Enforcer() *casbin.Enforcer {
	if c.enforcer == nil {
		var err error

		adapter, err := gormadapter.NewAdapterByDBUsePrefix(db.Get().DB(), "iac_")
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
