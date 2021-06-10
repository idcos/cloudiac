package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"

	"cloudiac/apps/task_manager"
	"cloudiac/cmds/common"
	"cloudiac/configs"
	"cloudiac/consts"
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/services"
	"cloudiac/utils/kafka"
	"cloudiac/utils/logs"
	"cloudiac/web"
)

type Option struct {
	common.OptionVersion

	Config     string `short:"c" long:"config"  default:"config.yml" description:"config file"`
	Verbose    []bool `short:"v" long:"verbose" description:"Show verbose debug message"`
	ReRegister bool   `long:"re-register" description:"Re registration service to Consul"`
}

var opt = Option{}

func main() {
	_, err := flags.Parse(&opt)
	if err != nil {
		os.Exit(1)
	}
	common.ShowVersionIf(opt.Version)

	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Panic(err)
		}
	} else {
		log.Println(err)
	}

	configs.Init(opt.Config)
	conf := configs.Get().Log
	logs.Init(conf.LogLevel, conf.LogPath, conf.LogMaxDays)

	// 依赖中间件及数据的初始化
	{
		db.Init()
		models.Init(true)

		tx := db.Get().Begin()
		defer func() {
			if r := recover(); r != nil {
				_ = tx.Rollback()
				panic(r)
			}
		}()
		// 自动执行平台初始化操作，只在第一次启动时执行
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

	return nil
}

// initAdmin 初始化 admin 账号
// 该函数读取环境变量 IAC_ADMIN_EMAIL、 IAC_ADMIN_PASSWORD 来获取初始用户的 email 和 password，
// 如果 IAC_ADMIN_EMAIL 未设置则使用默认邮箱,
func initAdmin(tx *db.Session) error {
	email := os.Getenv("IAC_ADMIN_EMAIL")
	password := os.Getenv("IAC_ADMIN_PASSWORD")

	if email == "" {
		email = consts.DefaultAdminEmail
	}

	// 通过邮箱查找账号，如果不存在则创建。
	// 如果用户修改环境变量 IAC_ADMIN_EMAIL 则会创建一个新用户
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
		OrgId:    0,
		Name:     "默认仓库",
		VcsType:  consts.GitTypeLocal,
		Status:   "enable",
		Address:  consts.LocalGitReposPath,
		VcsToken: "",
	}

	dbVcs := models.Vcs{}
	err := services.QueryVcs(0, "", "", tx).First(&dbVcs)
	if err != nil && !e.IsRecordNotFound(err) {
		return err
	}

	if dbVcs.Id == 0 { // 未创建
		_, err = services.CreateVcs(tx, vcs)
		if err != nil {
			return err
		}
	} else { // 己存在，进行更新
		vcs.Status = ""	// 不更新状态
		_, err = tx.Model(&vcs).Where("id = ?", dbVcs.Id).Update(vcs)
		if err != nil {
			return err
		}
	}
	return nil
}
