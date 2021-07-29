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

func UpdateSystemConfig(tx *db.Session, name string, attrs models.Attrs) (cfg *models.SystemCfg, re e.Error) {
	if name == models.SysCfgNameMaxJobsPerRunner {
		runnerMax, err := strconv.Atoi(attrs["value"].(string))
		if err != nil {
			return nil, e.New(e.BadRequest, fmt.Errorf("%s update err: %s", models.SysCfgNameMaxJobsPerRunner, err))
		}
		UpdateRunnerMax(runnerMax)
	}
	cfg = &models.SystemCfg{}
	if _, err := models.UpdateAttr(tx.Where("name = ?", name), &models.SystemCfg{}, attrs); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("update sys config error: %v", err))
	}

	if err := tx.Where("name = ?", name).First(cfg); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query sys config error: %v", err))
	}
	return
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
