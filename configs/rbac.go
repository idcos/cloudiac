package configs

import (
	"cloudiac/portal/libs/db"
	"cloudiac/utils/logs"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v2"
	"strings"
)

var RbacModel = `
[request_definition]
r = org, proj, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
# 按角色授权访问权限，动作为 * 的允许所有动作权限（适用于各类管理员）
# (r.org == p.sub || r.proj == p.sub) && r.obj == p.obj && (r.act == p.act || p.act == "*"))
# 平台管理员角色允许访问所有内容
# r.sub == "root"
# 允许资源匿名访问
# p.sub == "anonymous"

m = ((r.org == p.sub || r.proj == p.sub) && r.obj == p.obj && (r.act == p.act || p.act == "*")) || (p.sub == "anonymous" && r.obj == p.obj && (r.act == p.act || p.act == "*")) || r.org == "root"
`

// Policy 权限策略表
type Policy struct {
	sub string // 角色，包含组织角色和项目角色
	obj string // 资源名称
	act string // 动作，对资源的操作
}

// 角色策略
var polices = []Policy{
	// 配置方式：
	// 角色    资源    动作
	// 角色：
	//    组织角色：
	//    1. anonymous: （内置角色）非登陆用户
	//    2. root: （内置角色）平台管理员，通过 user.IsAdmin 指定
	//    3. login: （内置角色）登陆用户（无需组织信息）
	//    4. admin: 组织管理员
	//    5. member: 普通用户
	//    项目角色
	//    1. manager: 管理者
	//    2. approver: 审批者
	//    3. operator: 执行者
	//    4. guest: 访客
	// 资源
	//    对于访问的 URL  /api/v1/orgs/org-c3eotk06n88iemk0go90
	//    默认情况下解析出第三段 URL 的 orgs 作为资源名称
	//    如需重命名，可以在 router 通过中间件 a(obj, act) 调用方式重命名资源
	// 动作
	//    HTTP 访问对应如下：
	//    GET    read
	//    POST   create
	//    PUT    update
	//    PATCH  update
	//    DELETE delete
	//    需要替换或者扩展可以在 router 通过中间件 a(obj, act) 调用方式重命名动作
	//
	// 平台管理员（通过 rbac model 实现）
	// {"root", "*", "*"},
	// 匿名用户（跳过鉴权中间件）
	// {"anonymous", "auth", "read"},
	// 登陆用户
	{"login", "token", "*"},
	{"login", "runner", "*"},
	{"login", "consul", "*"},
	{"login", "webhook", "*"},

	// 用户
	{"admin", "users", "*"},
	{"member", "users", "read"},
	{"login", "self", "read/update"},
	{"admin", "self", "read/update"},
	{"member", "self", "read/update"},

	// 组织
	//{"root", "orgs", "*"},
	{"login", "orgs", "read"},
	{"admin", "orgs", "read/update"},
	{"admin", "orgs", "listuser/adduser/removeuser/updaterole"},
	{"member", "orgs", "read"},

	{"admin", "projects", "*"},
	{"member", "projects", "read"},

	{"admin", "templates", "*"},
	{"member", "templates", "read"},

	{"admin", "variables", "*"},
	{"member", "variables", "read"},

	// 项目
	{"manager", "projects", "*"},
	{"approver", "projects", "read"},
	{"operator", "projects", "read"},
	{"guest", "projects", "read"},

	{"manager", "envs", "*"},
	{"manager", "tasks", "*"},
	{"approver", "envs", "*"},
	{"approver", "tasks", "*"},
	{"operator", "envs", "read/update/deploy/destroy"},
	{"operator", "tasks", "read"},
	{"guest", "envs", "read"},
	{"guest", "tasks", "read"},

	{"manager", "templates", "*"},
	{"approver", "templates", "*"},
	{"operator", "templates", "read"},
	{"guest", "templates", "read"},

	// 变量
	{"admin", "variables", "*"},
	{"member", "variables", "read"},

	{"manager", "variables", "*"},
	{"approver", "variables", "*"},
	{"operator", "variables", "*"},
	{"guest", "variables", "read"},

	//token
	{"admin", "tokens", "*"},
	{"member", "tokens", "read"},

	{"manager", "tokens", "*"},
	{"approver", "tokens", "*"},
	{"operator", "tokens", "*"},
	{"guest", "tokens", "read"},

	//通知
	{"admin", "notifications", "*"},
	{"member", "notifications", "read"},

	//vcs
	{"admin", "vcs", "*"},
	{"member", "vcs", "read"},

	{"manager", "vcs", "read"},
	{"approver", "vcs", "read"},
	{"operator", "vcs", "read"},
	{"guest", "vcs", "read"},

	//runner
	{"manager", "runners", "read"},
	{"approver", "runners", "read"},
	{"operator", "runners", "read"},
	{"guest", "runners", "read"},

	// 密钥
	{"admin", "keys", "*"},
	{"member", "keys", "*"},

	// 演示模式，当访问演示组织下的资源，进入受限模式
	{"demo", "orgs", "read"},
	{"demo", "users", "read"},
	{"demo", "self", "read"},
	{"demo", "projects", "read"},
	{"demo", "tokens", "read"},
	{"demo", "notifications", "read"},
	{"demo", "vcs", "read"},
	{"demo", "runners", "read"},
	{"demo", "keys", "read"},
	{"demo", "templates", "read"},
	{"demo", "envs", "*"},
	{"demo", "tasks", "*"},
	{"demo", "variables", "*"},
}

// InitPolicy 初始化权限策略
func InitPolicy(tx *db.Session) error {
	logger := logs.Get().WithField("func", "initPolicy")
	logger.Infoln("init rbac policy...")
	var err error

	adapter, err := gormadapter.NewAdapterByDBUsePrefix(tx.DB(), "iac_")
	if err != nil {
		panic(fmt.Sprintf("error create enforcer: %v", err))
	}

	// 加载策略模型
	m, err := model.NewModelFromString(RbacModel)
	if err != nil {
		panic(fmt.Sprintf("error load rbac model: %v", err))
	}
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		panic(fmt.Sprintf("error create enforcer: %v", err))
	}

	// 初始化策略数据库
	for _, policy := range polices {
		for _, act := range strings.Split(policy.act, "/") {
			logger.Debugf("add policy: %s %s %s", policy.sub, policy.obj, act)
			enforcer.AddPolicy(policy.sub, policy.obj, act)
		}
	}

	return nil
}
