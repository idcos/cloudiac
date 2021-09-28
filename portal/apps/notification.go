// Copyright 2021 CloudJ Company Limited. All rights reserved.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"github.com/lib/pq"
	"strings"
)

type RespNotification struct {
	models.Notification
	EventType   string   `json:"-" form:"-" gorm:"event_type"`
	EventTypes  []string `json:"eventType" form:"eventType" gorm:"-"`
	CreatorName string   `json:"creatorName" form:"creatorName" `
}

func SearchNotification(c *ctx.ServiceContext, form *forms.SearchNotificationForm) (interface{}, e.Error) {
	notify := make([]*RespNotification, 0)
	query := services.SearchNotification(c.DB(), c.OrgId, c.ProjectId)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	if err := p.Scan(&notify); err != nil {
		return nil, e.New(e.DBError, err)
	}
	for index, v := range notify {
		notify[index].EventTypes = strings.Split(v.EventType, ",")
	}
	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     notify,
	}, nil
}

func DeleteNotification(c *ctx.ServiceContext, id models.Id) (result interface{}, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("Delete notification id: %s", id))
	err = services.DeleteNotification(c.DB(), id, c.OrgId)
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
		attrs["userIds"] = pq.StringArray(form.UserIds)
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

	if err := tx.Insert(&events); err != nil {
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

	tx := c.Tx()
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
		UserIds:   pq.StringArray(form.UserIds),
		Creator:   c.UserId,
	}, form.EventType)

	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return notification, nil
}

func DetailNotification(c *ctx.ServiceContext, form *forms.DetailNotificationForm) (interface{}, e.Error) {
	return services.DetailNotification(c.DB(), form.Id)
}
