package ctx

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
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

func (c *ServiceCtx) ProjectDB() *db.Session {
	return c.DB().Where("org_id = ? AND project_id = ?", c.OrgId, c.ProjectId)
}

func (c *ServiceCtx) Tx() *db.Session {
	return c.DB().Begin()
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

// RestrictOrg 获取组织数据隔离查询
// 带 IaC-Org-Id 只返回该 OrgId 相关数据集
// 不带 Iac-Org-Id:
// 1. 平台管理员返回所有组织数据
// 2. 其他用户返回该用户已经关联的组织的数据集
func (c *ServiceCtx) RestrictOrg(query *db.Session, m models.Modeler) *db.Session {
	query = query.Model(m)
	if c.OrgId != "" {
		query = query.Where(fmt.Sprintf("%s.org_id = ?", m.TableName()), c.OrgId)
	} else {
		// 如果是管理员，不需要附加限制参数，返回所有数据
		// 组织管理员或者普通用户，如果不带 org，应该返回该用户关联的所有 org
		if !c.IsSuperAdmin {
			subQ := query.Model(models.UserOrg{}).Select("org_id").Where("user_id = ?", c.UserId)
			query = query.Where(fmt.Sprintf("%s.org_id in (?)", m.TableName()), subQ.Expr())
		}
	}

	return query
}

// RestrictProject 获取项目数据隔离查询
// 如果带 IaC-Project-Id，返回该 ProjectId 相关数据
// 如果不带 IaC-Project-Id
// 1. 平台管理员返回所有数据
// 2. 如果没有 OrgId，返回该用户关联的所有项目数据
// 3. 如果有 OrgId，组织管理员返回该组织下的所有项目，普通用户返回该用户关联的项目
func (c *ServiceCtx) RestrictProject(query *db.Session, m models.Modeler) *db.Session {
	if c.ProjectId != "" {
		query = query.Where(fmt.Sprintf("%s.project_id = ?", m.TableName()), c.ProjectId)
	} else if c.IsSuperAdmin {
		// 平台管理员查询所有项目
	} else {
		// 如果没带 orgId, 访问用户关联的所有 project
		if c.OrgId == "" {
			subQ := c.DB().Model(models.UserProject{}).Select("project_id").Where("user_id = ?", c.UserId)
			query = query.Where(fmt.Sprintf("%s.project_id in (?)", m.TableName()), subQ.Expr())
		} else {
			// 如果带 orgId，如果是组织管理员，访问组织内所有 project
			// FIXME: 是否解析出了 role ?
			if c.Role == consts.OrgRoleAdmin {
				// 组织管理员，查询组织所有项目
				subQ := c.DB().Model(models.Project{}).Select("id").Where("org_id = ?", c.OrgId)
				query = query.Where(fmt.Sprintf("%s.project_id in (?)", m.TableName()), subQ.Expr())
			} else {
				// 普通用户，看用户在该组织关联了哪些项目
				subQ := c.DB().Model(models.Project{}).Select("id").Where("org_id = ?", c.OrgId).
					Joins(fmt.Sprintf("left join %s as o on o.project_id = %s.id where o.user_id = ?",
						models.UserProject{}.TableName(), models.Project{}.TableName()), c.UserId)
				query = query.Where(fmt.Sprintf("%s.project_id in (?)", m.TableName()), subQ.Expr())
			}
		}
	}

	return query
}
