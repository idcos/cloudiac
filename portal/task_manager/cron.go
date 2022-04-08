// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package task_manager

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/services"
	"cloudiac/portal/services/billcollect"
	"cloudiac/utils/logs"
	"context"
	"github.com/robfig/cron/v3"
	"time"
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
		}
	}()

	// 获取配置了账单采集的资源账号
	vg, err := services.GetVgByBillConf(tx)
	if err != nil {
		_ = tx.Rollback()
		logger.Errorf("get vg err: %s", err)
		return
	}

	for index, v := range vg {
		// 获取账单provider
		bp, err := billcollect.GetBillProvider(&vg[index])
		if err != nil {
			logger.Errorf("get bill provider failed vgId: %s provider: %s,err: %s", v.Id, v.Provider, err)
			continue
		}
		// 下载账单
		resCostAttr, resourceIds, err := bp.DownloadMonthBill(billingCycle)
		if err != nil {
			logger.Errorf("download bill failed vgId: %s provider: %s,err: %s", v.Id, v.Provider, err)
			continue
		}

		// 查询资源账号关联的项目
		projectIds, err := services.GetProjectIdsByVgId(tx, v.Id)
		if err != nil {
			logger.Errorf("query project ids failed vgId: %s provider: %s,err: %s", v.Id, v.Provider, err)
			continue
		}

		// 查询iac resource数据
		res, err := services.GetResourceByIdsInProvider(tx, resourceIds, projectIds, v)
		if err != nil {
			logger.Errorf("query iac resource failed vgId: %s provider: %s,err: %s", v.Id, v.Provider, err)
			continue
		}
		// 解析账单数据，构建入库数据
		bills,resIds:= services.ParseBill(resCostAttr, res, v.Id)
		if len(bills) == 0 {
			logger.Infof("resource not matched collect billing vgId: %s provider: %s,err: %s", v.Id, v.Provider, err)
			continue
		}

		// 删除上次采集的数据
		if err := services.DeleteResourceBill(tx, resIds, billingCycle); err != nil {
			logger.Errorf("del last bill data failed vgId: %s provider: %s,err: %s", v.Id, v.Provider, err)
			continue
		}

		if err := tx.Insert(bills); err != nil {
			logger.Errorf("bill insert failed vgId: %s provider: %s,err: %s", v.Id, v.Provider, err)
			continue
		}

	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		logger.Errorf("bill task db commit err: %s", err)
	}

	logger.Info("stop bill collect")
}
