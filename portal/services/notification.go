package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

type NotificationResp struct {
	Id        models.Id `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	EventType string    `json:"eventType"`
}

func ListNotificationCfgs(tx *db.Session, orgId models.Id) (interface{}, error) {
	users := make([]*NotificationResp, 0)
	err := tx.Table(models.User{}.TableName()).
		Select(fmt.Sprintf("%s.name, %s.email, n.id, n.event_type", models.User{}.TableName(), models.User{}.TableName())).
		Joins(fmt.Sprintf("right join %s as n on %s.id = n.user_id", models.NotificationCfg{}.TableName(), models.User{}.TableName())).
		Where(fmt.Sprintf("n.org_id = %d", orgId)).Debug().Find(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func UpdateNotificationCfg(tx *db.Session, id models.Id, attrs models.Attrs) (notificationCfg *models.NotificationCfg, err e.Error) {
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

func DeleteOrganizationCfg(tx *db.Session, id models.Id, orgId models.Id) e.Error {
	if _, err := tx.Where("id = ? AND org_id = ?", id, orgId).Delete(&models.NotificationCfg{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete notification cfg error: %v", err))
	}
	return nil
}

func FindOrganizationCfgByUserId(tx *db.Session, orgId models.Id, userId models.Id, eventType string) (bool, error) {
	return tx.Table(models.NotificationCfg{}.TableName()).
		Where("org_id = ? AND user_id = ? AND event_type = ?", orgId, userId, eventType).Exists()
}
