// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package configs

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
	Sub string // 角色，包含组织角色和项目角色
	Obj string // 资源名称
	Act string // 动作，对资源的操作
}

// Polices 角色策略
var Polices = []Policy{
	// 配置方式：
	// 角色    资源    动作
	// 角色：
	//    组织角色：
	//    1. anonymous: （内置角色）非登陆用户
	//    2. root: （内置角色）平台管理员，通过 user.IsAdmin 指定
	//    3. login: （内置角色）登陆用户（无需组织信息）
	//    4. admin: 组织管理员
	//    5. member: 普通用户
	//    6. complianceManager: 合规管理员
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
	{"login", "systems", "read"},
	{"login", "system_config", "read"},

	// 系统配置
	{"admin", "systems", "read"},
	{"admin", "system_config", "read"},
	{"member", "systems", "read"},
	{"member", "system_config", "read"},

	// 合规策略
	// 组织角色
	{"admin", "policies", "*"},
	{"member", "policies", "read"},
	{"complianceManager", "policies", "*"},

	// 项目角色
	{"manager", "policies", "suppress/enablescan/scan"},
	{"approver", "policies", "suppress/enablescan/scan"},
	{"operator", "policies", "suppress/scan"},
	{"manager", "policies", "read"},
	{"approver", "policies", "read"},
	{"operator", "policies", "read"},
	{"guest", "policies", "read"},

	// 用户
	{"admin", "users", "*"},
	{"member", "users", "read"},
	{"complianceManager", "users", "read"},
	{"login", "self", "read/update"},
	{"admin", "self", "read/update"},
	{"member", "self", "read/update"},
	{"complianceManager", "self", "read/update"},

	// 组织
	{"root", "orgs", "*"},
	{"login", "orgs", "read"},
	{"admin", "orgs", "read/update"},
	{"admin", "orgs", "listuser/adduser/removeuser/updaterole"},
	{"member", "orgs", "read/listuser/adduser"},
	// {"member", "orgs", "read"},
	{"complianceManager", "orgs", "read"},
	{"manager", "orgs", "listuser/adduser"},

	// 项目
	{"admin", "projects", "*"},
	{"member", "projects", "read"},
	{"complianceManager", "projects", "read"},
	{"manager", "projects", "*"},
	{"approver", "projects", "read"},
	{"operator", "projects", "read"},
	{"guest", "projects", "read"},

	// 环境
	{"manager", "envs", "*"},
	{"approver", "envs", "*"},
	{"operator", "envs", "read/update/deploy/destroy"},
	{"guest", "envs", "read"},

	// 任务
	{"manager", "tasks", "*"},
	{"approver", "tasks", "*"},
	{"operator", "tasks", "read/abort"},
	{"guest", "tasks", "read"},

	// 云模板
	{"admin", "templates", "*"},
	{"member", "templates", "read"},
	{"complianceManager", "templates", "read"},

	{"manager", "templates", "*"},
	{"approver", "templates", "*"},
	{"operator", "templates", "read"},
	{"guest", "templates", "read"},

	// 变量
	{"admin", "variables", "*"},
	{"member", "variables", "read"},
	{"complianceManager", "variables", "read"},

	{"manager", "variables", "*"},
	{"approver", "variables", "*"},
	{"operator", "variables", "*"},
	{"guest", "variables", "read"},

	// 资源账号(变量组)
	{"admin", "var_groups", "*"},
	{"member", "var_groups", "read"},
	{"complianceManager", "var_groups", "read"},

	//token
	{"admin", "tokens", "*"},
	{"complianceManager", "tokens", "*"},

	{"manager", "tokens", "*"},
	{"approver", "tokens", "*"},
	{"operator", "tokens", "*"},

	//通知
	{"admin", "notifications", "*"},
	{"member", "notifications", "read"},
	{"complianceManager", "notifications", "read"},

	//vcs
	{"admin", "vcs", "*"},
	{"member", "vcs", "read"},
	{"complianceManager", "vcs", "read"},

	{"manager", "vcs", "read"},
	{"approver", "vcs", "read"},
	{"operator", "vcs", "read"},
	{"guest", "vcs", "read"},

	//runner
	{"member", "runners", "read"},
	{"complianceManager", "runners", "read"},
	{"manager", "runners", "read"},
	{"approver", "runners", "read"},
	{"operator", "runners", "read"},
	{"guest", "runners", "read"},

	// 密钥
	{"admin", "keys", "*"},
	{"member", "keys", "*"},
	{"complianceManager", "keys", "*"},

	// Registry 配置
	{"admin", "system_config", "*"},
	{"member", "system_config", "read"},
	{"complianceManager", "system_config", "read"},

	{"manager", "system_config", "*"},
	{"approver", "system_config", "read"},
	{"operator", "system_config", "read"},
	{"guest", "system_config", "read"},

	// Registry
	{"manager", "registry", "*"},
	{"approver", "registry", "read"},
	{"operator", "registry", "read"},
	{"guest", "registry", "read"},

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
	{"demo", "policies", "read"},
	{"demo", "registry", "read"},
}
