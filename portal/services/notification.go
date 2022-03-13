// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
	"strings"
)

func SearchNotification(dbSess *db.Session, orgId, projectId models.Id) *db.Session {
	n := models.Notification{}.TableName()
	query := dbSess.Table(n).
		Joins(fmt.Sprintf("left join %s as ne on %s.id = ne.notification_id",
			models.NotificationEvent{}.TableName(), n)).
		Joins(fmt.Sprintf("left join %s as user on %s.creator = user.id",
			models.User{}.TableName(), n)).
		Where(fmt.Sprintf("%s.org_id = ?", n), orgId)
	if projectId != "" {
		query = query.Where(fmt.Sprintf("%s.project_id = ?", n), projectId)
	}
	return query.LazySelectAppend(fmt.Sprintf("%s.*", n), "group_concat(ne.event_type) as event_type").
		LazySelectAppend("user.name as creator_name").
		Group(fmt.Sprintf("%s.id", n))
}

func SearchNotifyEventType(dbSess *db.Session, notifyId models.Id) ([]string, e.Error) {
	events := make([]string, 0)
	if err := dbSess.Table(models.NotificationEvent{}.TableName()).
		Where("notification_id = ?", notifyId).
		Pluck("event_type", &events); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return events, nil
}

func UpdateNotification(tx *db.Session, id models.Id, attrs models.Attrs) (notificationCfg *models.Notification, err e.Error) {
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Notification{}, attrs); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("update notification cfg error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(&models.Notification{}); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query org error: %v", err))
	}
	return
}

func CreateNotification(tx *db.Session, notification models.Notification, eventType []string) (*models.Notification, e.Error) {
	if notification.Id == "" {
		notification.Id = models.NewId("notif-")
	}
	if err := models.Create(tx, &notification); err != nil {
		return nil, e.New(e.DBError, err)
	}

	events := make([]models.NotificationEvent, 0, len(eventType))
	for _, v := range eventType {
		events = append(events, models.NotificationEvent{
			NotificationId: notification.Id,
			EventType:      v,
		})
	}

	if err := tx.Insert(&events); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return &notification, nil
}

func DeleteNotificationEvent(tx *db.Session, nId models.Id) e.Error {
	if _, err := tx.Where("notification_id = ?", nId).Delete(&models.NotificationEvent{}); err != nil {
		return e.New(e.DBError, err)
	}

	return nil
}

func DeleteNotification(tx *db.Session, id models.Id, orgId models.Id) e.Error {
	if _, err := tx.Where("id = ? AND org_id = ?", id, orgId).Delete(&models.Notification{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete notification cfg error: %v", err))
	}

	if err := DeleteNotificationEvent(tx, id); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete notification cfg error: %v", err))
	}

	return nil
}

type RespDetailNotification struct {
	models.Notification
	EventType  string   `json:"-" `
	EventTypes []string `json:"eventType" gorm:"-"`
}

func DetailNotification(dbSess *db.Session, id models.Id) (interface{}, e.Error) {
	resp := RespDetailNotification{}
	if err := dbSess.Table(models.Notification{}.TableName()).
		Joins(fmt.Sprintf("left join %s as ne on %s.id = ne.notification_id",
			models.NotificationEvent{}.TableName(), models.Notification{}.TableName())).
		Where(fmt.Sprintf("%s.id = ?", models.Notification{}.TableName()), id).
		LazySelectAppend(fmt.Sprintf("%s.*", models.Notification{}.TableName())).
		LazySelectAppend("group_concat(ne.event_type) as event_type").
		Group(fmt.Sprintf("%s.id", models.Notification{}.TableName())).
		First(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	resp.EventTypes = strings.Split(resp.EventType, ",")
	return resp, nil
}
