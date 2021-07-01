package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

func CreateAccessToken(tx *db.Session, webhook models.TemplateAccessToken) (*models.TemplateAccessToken, e.Error) {
	if err := models.Create(tx, &webhook); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return &webhook, nil
}

func UpdateAccessToken(tx *db.Session, id models.Id, attrs models.Attrs) (*models.TemplateAccessToken, e.Error) {
	webhook := &models.TemplateAccessToken{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.TemplateAccessToken{}, attrs); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("update vcs error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(webhook); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query vcs error: %v", err))
	}
	return webhook, nil
}

func DeleteAccessToken(tx *db.Session, id models.Id) (interface{}, e.Error) {
	if _, err := tx.Where("id = ?", id).Delete(&models.TemplateAccessToken{}); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("delete vcs error: %v", err))
	}
	return nil, nil
}

func DetailAccessToken(tx *db.Session, id models.Id) (interface{}, e.Error) {
	accessToken := &models.TemplateAccessToken{}
	err := tx.Where("id = ?", id).First(accessToken)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	return accessToken, nil
}

func SearchAccessTokenByTplGuid(tx *db.Session, guid string) *db.Session {
	return tx.Model(&models.TemplateAccessToken{}).Where("tpl_guid = ?", guid)
}

func SearchAccessTokenByTplId(tx *db.Session, id models.Id) *db.Session {
	return tx.Model(&models.TemplateAccessToken{}).Where("tpl_id = ?", id)
}
