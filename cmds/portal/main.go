package main

import (
	"cloudiac/services"
	"fmt"
	"os"

	"cloudiac/cmds/common"
	"cloudiac/configs"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"cloudiac/web"

	"github.com/jessevdk/go-flags"
)

type Option struct {
	common.OptionVersion

	Config  string `short:"c" long:"config"  default:"config.yml" description:"config file"`
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug message"`
}

var opt = Option{}

func main() {
	_, err := flags.Parse(&opt)
	if err != nil {
		os.Exit(1)
	}
	common.ShowVersionIf(opt.Version)

	logs.Init(utils.LogLevel(len(opt.Verbose)))
	configs.Init(opt.Config)

	db.Init()
	models.Init(true)
	//rabbitmq.InitAMQP(configs.Get().Rmq.Addr)

	logger := logs.Get()
	tx := db.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			logger.Fatalln(r)
		}
	}()

	// 自动执行平台初始化操作，只在第一次启动时执行
	if err := appAutoInit(tx); err != nil {
		panic(err)
	}
	if err := tx.Commit(); err != nil {
		logger.Fatalln(err)
	}

	web.StartServer()
}

// 平台初始化
func appAutoInit(tx *db.Session) (err error) {
	logger := logs.Get().WithField("func", "appAutoInit")
	logger.Infoln("running")

	// dev init
	admin, _ := services.GetUserByEmail(tx, "admin")
	if admin != nil {
		return nil
	}

	initPass := "Yunjikeji"
	hashedPassword, err := services.HashPassword(initPass)
	if err != nil {
		fmt.Println("111", err)
	}
	_, err = services.CreateUser(tx, models.User{
		Name:     "admin",
		Password: hashedPassword,
		Phone:    "",
		Email:    "admin",
		InitPass: initPass,
		IsAdmin: true,
	})
	if err != nil {
		return err
	}

	logger.Infoln("initialize ...")

	return nil
}
