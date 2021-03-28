package services

import (
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"fmt"
)

func ListNotificationCfgs(tx *db.Session, orgId uint) (interface{}, error) {
	users := []*models.User{}
	err := tx.Table(models.User{}.TableName()).
		Joins(fmt.Sprintf("right join %s as n on %s.id = n.user_id", models.NotificationCfg{}.TableName(), models.User{}.TableName())).
		Where(fmt.Sprintf("n.org_id = %d", orgId)).Debug().Find(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
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

func DeleteOrganizationCfg(tx *db.Session, orgId uint, userId uint) e.Error {
	if _, err := tx.Where("org_id = ? AND user_id = ?", orgId, userId).Debug().Delete(&models.NotificationCfg{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete notification cfg error: %v", err))
	}
	return nil
}
