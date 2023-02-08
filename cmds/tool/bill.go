// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package main

import (
	"cloudiac/configs"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/utils/logs"
	"time"
)

type BillCmd struct {
	OrgId     string `long:"orgId" short:"o" description:"billing collect with orgId" required:"false"`
	ProjectId string `long:"projectId" short:"p" description:"billing collect with projectId" required:"false"`
	VgId      string `long:"vgId" short:"v" description:"billing collect with vgId" required:"false"`
	Cycle     string `long:"cycle" short:"c" description:"billing collect cycle" required:"false"`
}

func (*BillCmd) Usage() string {
	return `<bill collect>`
}

func (b *BillCmd) Execute(args []string) error {
	var billingCycle string = b.Cycle

	configs.Init(opt.Config)
	db.Init(configs.Get().Mysql)
	models.Init(false)

	logger := logs.Get().WithField("acton", "billing cron task")
	logger.Info("start bill collect")

	if billingCycle == "" {
		billingCycle = time.Now().Format("2006-01")
	}

	tx := db.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	vgs, err := buildVgs(tx, b.OrgId, b.ProjectId, b.VgId)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// 去重
	vgm := make(map[string]interface{})
	newVgs := make([]models.VariableGroup, 0)
	for _, v := range vgs {
		if _, ok := vgm[v.Id.String()]; ok {
			continue
		}
		newVgs = append(newVgs, v)
	}

	for index := range newVgs {
		services.BuildVgBilling(tx, newVgs[index], logger, billingCycle)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		logger.Errorf("bill task db commit err: %s", err)
	}

	logger.Info("stop bill collect")
	return nil
}

func buildVgs(tx *db.Session, orgId, projectId, vgId string) ([]models.VariableGroup, error) {
	vgs := make([]models.VariableGroup, 0)
	if orgId != "" {
		orgVgs, err := services.GetVgByOrgId(tx, orgId)
		if err != nil {
			logger.Errorf("get org vg err: %s", err)
			return nil, err
		}
		vgs = append(vgs, orgVgs...)
	}

	if projectId != "" {
		orgVgs, err := services.GetVgByProjectId(tx, projectId)
		if err != nil {
			logger.Errorf("get project vg err: %s", err)
			return nil, err
		}
		vgs = append(vgs, orgVgs...)
	}

	if vgId != "" {
		orgVgs, err := services.GetVgById(tx, vgId)
		if err != nil {
			logger.Errorf("get vg err: %s", err)
			return nil, err
		}
		vgs = append(vgs, orgVgs...)
	}

	if orgId == "" && projectId == "" && vgId == "" {
		confVgs, err := services.GetVgByBillConf(tx)
		if err != nil {
			logger.Errorf("get vg err: %s", err)
			return nil, err
		}
		vgs = append(vgs, confVgs...)
	}
	return vgs, nil
}
