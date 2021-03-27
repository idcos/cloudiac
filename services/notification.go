package services

import (
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"fmt"
)

func ListNotificationCfgs(tx *db.Session, orgId int) (interface{}, error) {
	cfgs := []*models.NotificationCfg{}
	if err := tx.Where("org_id = ?", orgId).First(&cfgs); err != nil {
		return nil, err
	}
	return cfgs, nil
}

func UpdateNotificationCfg(tx *db.Session, id uint, attrs models.Attrs) (notificationCfg *models.NotificationCfg, err e.Error) {
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.NotificationCfg{}, attrs); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("update notification cfg error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(&models.NotificationCfg{}); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query org error: %v", err))
	}
	return
}

func CreateNotificationCfg(tx *db.Session, cfg models.NotificationCfg) (*models.NotificationCfg, e.Error) {
	if err := models.Create(tx, &cfg); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return &cfg, nil
}

func DeleteOrganizationCfg(tx *db.Session, cfgId int) e.Error {
	if _, err := tx.Where("id = ?", cfgId).Delete(&models.NotificationCfg{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete notification cfg error: %v", err))
	}
	return nil
}
