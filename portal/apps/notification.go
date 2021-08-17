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

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("notificationType") {
		attrs["notificationType"] = form.NotificationType
	}

	if form.HasKey("eventFailed") {
		attrs["eventFailed"] = form.EventFailed
	}

	if form.HasKey("eventComplete") {
		attrs["eventComplete"] = form.EventComplete
	}

	if form.HasKey("eventApproving") {
		attrs["eventApproving"] = form.EventApproving
	}

	if form.HasKey("eventRunning") {
		attrs["eventRunning"] = form.EventRunning
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

	cfg, err = services.UpdateNotificationCfg(c.DB(), form.Id, attrs)
	return cfg, err
}

func CreateNotification(c *ctx.ServiceContext, form *forms.CreateNotificationForm) (*models.Notification, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create org notification cfg %s", form.NotificationType))

	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	notificationCfg, err := services.CreateNotificationCfg(tx, models.Notification{
		OrgId:            c.OrgId,
		ProjectId:        c.ProjectId,
		Name:             form.Name,
		NotificationType: form.NotificationType,
		Secret:           form.Secret,
		Url:              form.Url,
		UserIds:          form.UserIds,
		EventFailed:      form.EventFailed,
		EventComplete:    form.EventComplete,
		EventApproving:   form.EventApproving,
		EventRunning:     form.EventRunning,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, err

	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return notificationCfg, nil
}
