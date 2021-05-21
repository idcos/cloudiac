package services

import (
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"fmt"
)

func CreateWebhook(tx *db.Session, webhook models.TemplateWebhook) (*models.TemplateWebhook, e.Error) {
	if err := models.Create(tx, &webhook); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return &webhook, nil
}

func UpdateWebhook(tx *db.Session, id uint, attrs models.Attrs) (*models.TemplateWebhook, e.Error) {
	webhook := &models.TemplateWebhook{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.TemplateWebhook{}, attrs); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("update vcs error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(webhook); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query vcs error: %v", err))
	}
	return webhook, nil
}

func DeleteWebhook(tx *db.Session, id uint) (interface{}, e.Error) {
	if _, err := tx.Where("id = ?", id).Delete(&models.TemplateWebhook{}); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("delete vcs error: %v", err))
	}
	return nil, nil
}

func DetailWebhook(tx *db.Session, id uint) (interface{}, e.Error) {
	webhook := &models.TemplateWebhook{}
	err := tx.Where("id = ?", id).First(webhook)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	return webhook, nil
}

func SearchWebhookByTplGuid(tx *db.Session, guid string) *db.Session {
	return tx.Model(&models.TemplateWebhook{}).Where("tpl_guid = ?", guid)
}

func SearchWebhookByTplId(tx *db.Session, id uint) *db.Session {
	return tx.Model(&models.TemplateWebhook{}).Where("tpl_id = ?", id)
}
