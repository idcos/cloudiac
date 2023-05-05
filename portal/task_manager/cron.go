// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package task_manager

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/services"
	"cloudiac/utils/logs"
	"context"
	"time"

	"github.com/robfig/cron/v3"
)

func billCron(ctx context.Context) {
	c := cron.New()
	if _, err := c.AddFunc("@daily", cronBillCollectTask); err != nil {
		logs.Get().Error("bill cron task start failed")
		return
	}
	c.Start()

	go func() {
		<-ctx.Done()
		c.Stop()
	}()
}

func cronBillCollectTask() {
	logger := logs.Get().WithField("action", "billing cron task")
	logger.Info("start bill collect")

	billingCycle := time.Now().Format("2006-01")

	tx := db.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 获取配置了账单采集的资源账号
	vgs, err := services.GetVgByBillConf(tx)
	if err != nil {
		_ = tx.Rollback()
		logger.Errorf("get vg err: %s", err)
		return
	}

	for index := range vgs {
		services.BuildVgBilling(tx, vgs[index], logger, billingCycle)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		logger.Errorf("bill task db commit err: %s", err)
	}

	logger.Info("stop bill collect")
}
