package main

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v2"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"

	"cloudiac/cmds/common"
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/portal/services/sshkey"
	"cloudiac/portal/web"
	"cloudiac/utils/kafka"
	"cloudiac/utils/logs"
)

type Option struct {
	common.OptionVersion

	Config     string `short:"c" long:"config"  default:"config-portal.yml" description:"config file"`
	Verbose    []bool `short:"v" long:"verbose" description:"Show verbose debug message"`
	ReRegister bool   `long:"re-register" description:"Re registration service to Consul"`
}

var opt = Option{}

func main() {
	common.LoadDotEnv()

	_, err := flags.Parse(&opt)
	if err != nil {
		os.Exit(1)
	}
	common.ShowVersionIf(opt.Version)

	configs.Init(opt.Config)
	conf := configs.Get().Log
	logs.Init(conf.LogLevel, conf.LogPath, conf.LogMaxDays)

	if err := initSSHKeyPair(); err != nil {
		panic(errors.Wrap(err, "init ssh key pair"))
	}

	// 依赖中间件及数据的初始化
	{
		db.Init(configs.Get().Mysql)
		models.Init(true)

		tx := db.Get().Begin()
		defer func() {
			if r := recover(); r != nil {
				_ = tx.Rollback()
				panic(r)
			}
		}()
		// 自动执行平台初始化操作
		if err := appAutoInit(tx); err != nil {
			panic(err)
		}
		if err := tx.Commit(); err != nil {
			panic(err)
		}

		services.MaintenanceRunnerPerMax()
		kafka.InitKafkaProducerBuilder()
	}

	// 注册到 consul
	common.ReRegisterService(opt.ReRegister, "IaC-Portal")

	// 启动后台 worker
	//go task_manager.Start(configs.Get().Consul.ServiceID)

	// 启动 web server
	web.StartServer()
}

// 平台初始化
func appAutoInit(tx *db.Session) (err error) {
	logger := logs.Get().WithField("func", "appAutoInit")
	logger.Infoln("initialize ...")

	if err := initAdmin(tx); err != nil {
		return errors.Wrap(err, "init admin account")
	}

	if err := initSystemConfig(tx); err != nil {
		return errors.Wrap(err, "init system config")
	}

	if err := initVcs(tx); err != nil {
		return errors.Wrap(err, "init vcs")
	}

	if err := initTemplates(tx); err != nil {
		return errors.Wrap(err, "init meat template")
	}

	if err := initPolicy(tx); err != nil {
		return errors.Wrap(err, "init rbac policy")
	}

	return nil
}

// initAdmin 初始化 admin 账号
// 该函数读取环境变量 IAC_ADMIN_EMAIL、 IAC_ADMIN_PASSWORD 来获取初始用户的 email 和 password，
// 如果 IAC_ADMIN_EMAIL 未设置则使用默认邮箱, IAC_ADMIN_PASSWORD 未设置则报错
func initAdmin(tx *db.Session) error {
	if ok, err := services.QueryUser(tx).Exists(); err != nil {
		return err
	} else if ok { // 己存在用户，跳过
		return nil
	}

	email := os.Getenv("IAC_ADMIN_EMAIL")
	password := os.Getenv("IAC_ADMIN_PASSWORD")

	if email == "" {
		email = consts.DefaultAdminEmail
	}

	// 通过邮箱查找账号，如果不存在则创建。
	admin, err := services.GetUserByEmail(tx, email)
	if err != nil && !e.IsRecordNotFound(err) {
		return err
	} else if admin != nil { // 用户己存在
		return nil
	}

	if password == "" {
		return fmt.Errorf("environment variable 'IAC_ADMIN_PASSWORD' is not set")
	}

	hashedPassword, err := services.HashPassword(password)
	if err != nil {
		return err
	}

	logger := logs.Get()
	logger.Infof("create admin account, email: %s", email)
	_, err = services.CreateUser(tx, models.User{
		Name:     email,
		Password: hashedPassword,
		Phone:    "",
		Email:    email,
		IsAdmin:  true,
	})
	return err
}

func initSystemConfig(tx *db.Session) (err error) {
	logger := logs.Get().WithField("func", "initSystemConfig")
	logger.Infoln("init system config...")

	initSysConfigs := []models.SystemCfg{
		{
			Name:        models.SysCfgNameMaxJobsPerRunner,
			Value:       "100",
			Description: "每个CT-Runner同时启动的最大容器数",
		}, {

			Name:        models.SysCfgNamePeriodOfLogSave,
			Value:       "Permanent",
			Description: "日志保存周期",
		},
	}

	tx = tx.Model(&models.SystemCfg{})
	for _, c := range initSysConfigs {
		if ok, err := tx.Where("name = ?", c.Name).Exists(); err != nil {
			return err
		} else if !ok {
			_, err = services.CreateSystemConfig(tx, c)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func initVcs(tx *db.Session) error {
	vcs := models.Vcs{
		OrgId:    "",
		Name:     "默认仓库",
		VcsType:  consts.GitTypeLocal,
		Status:   "enable",
		Address:  consts.LocalGitReposPath,
		VcsToken: "",
	}

	dbVcs := models.Vcs{}
	err := services.QueryVcs("", "", "", tx).First(&dbVcs)
	if err != nil && !e.IsRecordNotFound(err) {
		return err
	}

	if dbVcs.Id == "" { // 未创建
		_, err = services.CreateVcs(tx, vcs)
		if err != nil {
			return err
		}
	} else { // 己存在，进行更新
		vcs.Status = "" // 不更新状态
		_, err = tx.Model(&vcs).Where("id = ?", dbVcs.Id).Update(vcs)
		if err != nil {
			return err
		}
	}
	return nil
}

func initTemplates(tx *db.Session) error {
	// TODO: 导入内置演示模板
	return nil
}

func initSSHKeyPair() error {
	return sshkey.InitSSHKeyPair()
}

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
	//    1. owner: 项目管理员
	//    2. manager: 审批者
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
	{"login", "system", "*"},
	{"login", "runner", "*"},
	{"login", "consul", "*"},
	{"login", "webhook", "*"},
	{"login", "self", "read/update"},

	// 组织
	//{"root", "orgs", "*"},
	{"admin", "orgs", "read/update/delete"},
	{"admin", "orgs", "listuser/adduser/removeuser/updaterole"},
	{"member", "orgs", "read"},
	{"admin", "users", "*"},
	{"member", "users", "read"},

	// 项目
	{"owner", "projects", "*"},
	{"manager", "projects", "read"},
	{"operator", "projects", "read"},
	{"guest", "projects", "read"},

	{"owner", "envs", "*"},
	{"manager", "envs", "*"},
	{"operator", "envs", "read/update/deploy/deleteres"},
	{"guest", "envs", "read"},

	{"owner", "templates", "*"},
	{"manager", "templates", "*"},
	{"operator", "templates", "read"},
	{"guest", "templates", "read"},
}

// initPolicy 初始化权限策略
func initPolicy(tx *db.Session) error {
	logger := logs.Get().WithField("func", "initPolicy")
	logger.Infoln("init rbac policy...")
	var err error

	adapter, err := gormadapter.NewAdapterByDBUsePrefix(tx.DB(), "iac_")
	if err != nil {
		panic(fmt.Sprintf("error create enforcer: %v", err))
	}

	// 加载策略模型
	// TODO: 模型初始化到数据库，减少外部文件
	enforcer, err := casbin.NewEnforcer("configs/rbac_model.conf", adapter)
	if err != nil {
		panic(fmt.Sprintf("error create enforcer: %v", err))
	}

	for _, policy := range polices {
		for _, act := range strings.Split(policy.act, "/") {
			logger.Debugf("add policy: %s %s %s", policy.sub, policy.obj, act)
			enforcer.AddPolicy(policy.sub, policy.obj, act)
		}
	}

	return nil
}
