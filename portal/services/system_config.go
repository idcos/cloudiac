// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
	"strconv"
)

func QuerySystemConfig(dbSess *db.Session) *db.Session {
	return dbSess.Model(&models.SystemCfg{})
}

func UpdateSystemConfig(tx *db.Session, name string, attrs models.Attrs) (int64, e.Error) {
	if name == models.SysCfgNameMaxJobsPerRunner {
		runnerMax, err := strconv.Atoi(attrs["value"].(string))
		if err != nil {
			return 0, e.New(e.BadRequest, fmt.Errorf("%s update err: %s", models.SysCfgNameMaxJobsPerRunner, err))
		}
		UpdateRunnerMax(runnerMax)
	}
	if changed, err := models.UpdateAttr(tx.Where("name = ?", name), &models.SystemCfg{}, attrs); err != nil {
		return 0, e.New(e.DBError, fmt.Errorf("update sys config error: %v", err))
	} else {
		return changed, nil
	}
}

func CreateSystemConfig(tx *db.Session, cfg models.SystemCfg) (*models.SystemCfg, e.Error) {
	if cfg.Id == "" {
		cfg.Id = models.NewId("")
	}
	if err := models.Create(tx, &cfg); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return &cfg, nil
}

func GetMigrationVersion(tx *db.Session) (string, error) {
	c := models.SystemCfg{}
	err := QuerySystemConfig(tx).Where("name = ?", models.SysCfgNameMigrationVersion).First(&c)
	if e.IsRecordNotFound(err) {
		err = nil
	}
	return c.Value, err
}

func SaveMigrationVersion(tx *db.Session, ver string) error {
	changed, err := UpdateSystemConfig(tx, models.SysCfgNameMigrationVersion, models.Attrs{"value": ver})
	if e.IsRecordNotFound(err) || changed == 0 {
		_, err = CreateSystemConfig(tx, models.SystemCfg{
			Name:        models.SysCfgNameMigrationVersion,
			Value:       ver,
			Description: "",
		})
	}
	return err
}
