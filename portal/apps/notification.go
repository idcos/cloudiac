// Copyright 2021 CloudJ Company Limited. All rights reserved.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
)

func SearchNotification(c *ctx.ServiceContext) (interface{}, e.Error) {
	cfgs, err := services.SearchNotification(c.DB(), c.OrgId)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	return cfgs, nil
}

func DeleteNotification(c *ctx.ServiceContext, id models.Id) (result interface{}, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("Delete org notification id: %s", id))
	err = services.DeleteOrganizationCfg(c.DB(), id, c.OrgId)
	if err != nil {
		return nil, err
	}
	return
}

func UpdateNotification(c *ctx.ServiceContext, form *forms.UpdateNotificationForm) (cfg *models.Notification, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update org notification cfg id: %s", form.Id))

	if form.Id == "" {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}

	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("type") {
		attrs["type"] = form.Type
	}

	if form.HasKey("secret") {
		attrs["secret"] = form.Secret
	}

	if form.HasKey("url") {
		attrs["url"] = form.Url
	}

	if form.HasKey("userIds") {
		attrs["userIds"] = form.UserIds
	}

	cfg, err = services.UpdateNotification(tx, form.Id, attrs)
	if err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	if err := services.DeleteNotificationEvent(tx, form.Id); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	events := make([]models.NotificationEvent, len(form.EventType))
	for _, v := range form.EventType {
		events = append(events, models.NotificationEvent{
			NotificationId: form.Id,
			EventType:      v,
		})
	}

	if err := services.CreateNotificationEvent(tx, events); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return cfg, err
}

func CreateNotification(c *ctx.ServiceContext, form *forms.CreateNotificationForm) (*models.Notification, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create org notification cfg %s", form.Type))

	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	notification, err := services.CreateNotification(tx, models.Notification{
		OrgId:     c.OrgId,
		ProjectId: c.ProjectId,
		Name:      form.Name,
		Type:      form.Type,
		Secret:    form.Secret,
		Url:       form.Url,
		UserIds:   form.UserIds,
	})

	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	events := make([]models.NotificationEvent, len(form.EventType))
	for _, v := range form.EventType {
		events = append(events, models.NotificationEvent{
			NotificationId: notification.Id,
			EventType:      v,
		})
	}

	if err := services.CreateNotificationEvent(tx, events); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return notification, nil
}
