package main

import (
	task_manager2 "cloudiac/apps/task_manager"
	"cloudiac/cmds/common"
	"cloudiac/configs"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/services"
	"cloudiac/utils/kafka"
	"cloudiac/utils/logs"
	"cloudiac/web"
	//_ "net/http/pprof"
	"os"

	"github.com/jessevdk/go-flags"
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

	configs.Init(opt.Config)
	conf := configs.Get().Log
	logs.Init(conf.LogLevel, conf.LogMaxDays, "iac-portal")
	common.ReRegisterService(opt.ReRegister, "IaC-Portal")

	db.Init()
	models.Init(true)

	logger := logs.Get()
	tx := db.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			logger.Fatalln(r)
		}
	}()

	// 自动执行平台初始化操作，只在第一次启动时执行
	InitVcs(tx)
	if err := appAutoInit(tx); err != nil {
		panic(err)
	}
	if err := tx.Commit(); err != nil {
		logger.Fatalln(err)
	}
	kafka.InitKafkaProducerBuilder()
	//go services.RunTaskToRunning()
	//go services.RunTaskState()
	services.MaintenanceRunnerPerMax()

	//go services.RunTask()

	go task_manager2.Start(configs.Get().Consul.ServiceID)

	//go http.ListenAndServe("0.0.0.0:6060", nil)
	web.StartServer()
}

// InitVcs TODO: 默认使用内置 http git server
func InitVcs(tx *db.Session) {
	logger := logs.Get()
	vcs := models.Vcs{
		OrgId:    0,
		Name:     "默认仓库",
		VcsType:  configs.Get().Gitlab.Type,
		Status:   "enable",
		Address:  configs.Get().Gitlab.Url,
		VcsToken: configs.Get().Gitlab.Token,
	}
	exist, err := services.QueryVcs(0, "", "", tx).Exists()
	if err != nil {
		logger.Error(err)
		return
	}
	if !exist {
		_, err = services.CreateVcs(tx, vcs)
		if err != nil {
			logger.Error(err)
		}
	} else {
		attrs := models.Attrs{
			"vcs_type":  configs.Get().Gitlab.Type,
			"address":   configs.Get().Gitlab.Url,
			"vcs_token": configs.Get().Gitlab.Token,
		}
		_, err = services.UpdateVcs(tx, 0, attrs)
		if err != nil {
			logger.Error(err)
		}
	}
}

// 平台初始化
func appAutoInit(tx *db.Session) (err error) {
	logger := logs.Get().WithField("func", "appAutoInit")
	logger.Infoln("running")

	// dev init
	if err := initAdmin(tx); err != nil {
		return err
	}

	if err := initSystemConfig(tx); err != nil {
		return err
	}

	logger.Infoln("initialize ...")

	return nil
}

func initAdmin(tx *db.Session) (err error) {
	log := logs.Get()
	admin, _ := services.GetUserByEmail(tx, "admin")
	if admin != nil {
		return nil
	}

	initPass := "Yunjikeji"
	hashedPassword, err := services.HashPassword(initPass)
	if err != nil {
		log.Errorf("init admin err: %+v", err)
	}
	_, err = services.CreateUser(tx, models.User{
		Name:     "admin",
		Password: hashedPassword,
		Phone:    "",
		Email:    "admin",
		InitPass: initPass,
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
