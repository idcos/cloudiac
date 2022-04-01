package task_manager

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/portal/services/billcollect"
	"cloudiac/utils/logs"
	"context"
	"github.com/robfig/cron/v3"
	"time"
)

func billCron(ctx context.Context) {
	f := func() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		cronTask()
	}

	c := cron.New()
	if _, err := c.AddFunc("@every 1s", f); err != nil {
		logs.Get().Error("bill cron task start failed")
		return
	}
	c.Start()
}

func cronTask() {
	logger := logs.Get().WithField("bill", "cron task")
	tx := db.Get().Begin()
	billingCycle:=time.Now()
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

	resources := make([]models.Resource, 0)
	billData := make([]billcollect.ResourceCost,0)
	for _, v := range vg {
		// 获取资源账号下的资源
		rs, err := services.GetResourceByVg(tx, v)
		if err != nil {
			logger.Errorf("get resource failed vg: %+v,err: %s", v, err)
			continue
		}
		resources = append(resources, rs...)

		// 获取账单数据
		bd,err:=services.BillData(v,billingCycle)
		if err!= nil{
			logger.Errorf("get resource failed vg: %+v,err: %s", v, err)
			continue
		}
		billData = append(billData,bd...)

	}
	// 数据匹配入库
}
