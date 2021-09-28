// Copyright 2021 CloudJ Company Limited. All rights reserved.

package main

import (
	common2 "cloudiac/common"
	"cloudiac/portal/apps"
	"cloudiac/portal/task_manager"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"os"

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

	//if err := initSSHKeyPair(); err != nil {
	//	panic(errors.Wrap(err, "init ssh key pair"))
	//}

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
	go task_manager.Start(configs.Get().Consul.ServiceID)

	// 获取演示组织ID
	org, _ := services.GetDemoOrganization(db.Get())
	if org != nil {
		common2.DemoOrgId = org.Id.String()
	}
	// 初始化tfversions list
	apps.InitTfVersions()
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

	if err := initSysUser(tx); err != nil {
		return errors.Wrap(err, "init sys account")
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
	if err != nil && err.Code() != e.UserNotExists {
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

// initSysUser 初始化 sys 账号
// 该函数读取环境变量 IAC_SYS_EMAIL, IAC_SYS_NAME 来获取初始系统用户的 email 和用户名
// 如果 IAC_SYS_EMAIL 未设置则使用 DefaultSysEmail 邮箱，IAC_SYS_NAME 未设置则使用默认系统用户名
func initSysUser(tx *db.Session) error {
	if ok, err := services.QueryUser(tx).Where("id = ?", consts.SysUserId).Exists(); err != nil {
		return err
	} else if ok { // 己存在用户，跳过
		return nil
	}

	email := os.Getenv("IAC_SYS_EMAIL")
	if email == "" {
		email = consts.DefaultSysEmail
	}
	name := os.Getenv("IAC_SYS_NAME")
	if name == "" {
		name = consts.DefaultSysName
	}

	// 通过邮箱查找账号，如果不存在则创建。
	sys, err := services.GetUserByEmail(tx, email)
	if err != nil && err.Code() != e.UserNotExists {
		return err
	} else if sys != nil { // 用户己存在
		return fmt.Errorf("sys email conflict")
	}

	logger := logs.Get()
	logger.Infof("create sys account, email: %s, name: %s", email, name)
	u := models.User{
		Name:  name,
		Phone: "",
		Email: email,
	}
	u.Id = consts.SysUserId
	_, err = services.CreateUser(tx, u)
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

// initVcs 初始化系统默认 VCS
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
	err := services.QueryVcs("", "", "", true, tx).First(&dbVcs)
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
