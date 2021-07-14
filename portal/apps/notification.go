package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"encoding/json"
	"fmt"
)

func ListNotificationCfgs(c *ctx.ServiceCtx) (interface{}, e.Error) {
	cfgs, err := services.ListNotificationCfgs(c.DB(), c.OrgId)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	return cfgs, nil
}

func DeleteNotificationCfg(c *ctx.ServiceCtx, id models.Id) (result interface{}, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("Delete org notification id: %s", id))
	err = services.DeleteOrganizationCfg(c.DB(), id, c.OrgId)
	if err != nil {
		return nil, err
	}
	return
}

func UpdateNotificationCfg(c *ctx.ServiceCtx, form *forms.UpdateNotificationCfgForm) (cfg *models.NotificationCfg, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update org notification cfg id: %s", form.NotificationId))

	if form.NotificationId == "" {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}

	attrs := models.Attrs{}
	if form.HasKey("notificationType") {
		attrs["notificationType"] = form.NotificationType
	}

	if form.HasKey("eventType") {
		attrs["eventType"] = form.EventType
	}

	if form.HasKey("cfgInfo") {
		cfgInfo := form.CfgInfo
		cfgJson, _ := json.Marshal(cfgInfo)
		attrs["cfgInfo"] = cfgJson
	}

	cfg, err = services.UpdateNotificationCfg(c.DB(), form.NotificationId, attrs)
	return cfg, err
}

func CreateNotificationCfg(c *ctx.ServiceCtx, form *forms.CreateNotificationCfgForm) (*models.NotificationCfg, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create org notification cfg %s", form.NotificationType))

	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	notificationCfg, err := func() (*models.NotificationCfg, e.Error) {
		var (
			notificationCfg *models.NotificationCfg
			err             e.Error
		)

		cfgInfo := form.CfgInfo
		cfgJson, _ := json.Marshal(cfgInfo)

		if form.NotificationType == "email" {
			for _, userId := range form.UserIds {
				isExists, _ := services.FindOrganizationCfgByUserId(tx, c.OrgId, userId, form.EventType)
				if isExists {
					continue
				}
				notificationCfg, err = services.CreateNotificationCfg(tx, models.NotificationCfg{
					OrgId:            c.OrgId,
					NotificationType: form.NotificationType,
					EventType:        form.EventType,
					UserId:           userId,
				})
				if err != nil {
					if e.IsDuplicate(err) {
						continue
					} else {
						return nil, err
					}
				}
			}
		} else {
			notificationCfg, err = services.CreateNotificationCfg(tx, models.NotificationCfg{
				OrgId:            c.OrgId,
				NotificationType: form.NotificationType,
				EventType:        form.EventType,
				CfgInfo:          cfgJson,
			})
			if err != nil {
				return nil, err
			}
		}

		return notificationCfg, nil
	}()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return notificationCfg, nil
}
