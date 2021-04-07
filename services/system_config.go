package services

import (
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"fmt"
)

func QuerySystemConfig(query *db.Session) *db.Session {
	return query.Model(&models.SystemCfg{}).
		Where("name = 'MAX_JOBS_PER_RUNNER'")
}

func UpdateSystemConfig(tx *db.Session, attrs models.Attrs) (cfg *models.SystemCfg, re e.Error) {
	cfg = &models.SystemCfg{}
	if _, err := models.UpdateAttr(tx.Where("name = 'MAX_JOBS_PER_RUNNER'"), &models.SystemCfg{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.OrganizationAliasDuplicate)
		}
		return nil, e.New(e.DBError, fmt.Errorf("update sys config error: %v", err))
	}
	if err := tx.Where("name = 'MAX_JOBS_PER_RUNNER'").First(cfg); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query sys config error: %v", err))
	}
	return
}

func CreateSystemConfig(tx *db.Session, cfg models.SystemCfg) (*models.SystemCfg, e.Error) {
	if err := models.Create(tx, &cfg); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return &cfg, nil
}
