package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

func CreateTag(session *db.Session, tag []models.Tag) e.Error {
	if err := models.CreateBatch(session, tag); err != nil {
		return e.New(e.DBError, err)
	}

	return nil
}

func SearchTag(session *db.Session, objectId models.Id, objectType string) (tags []models.Tag, error e.Error) {
	if err := session.Model(models.Tag{}).
		Where("object_id = ?", objectId).
		Where("object_type = ?", objectType).
		Find(&tags); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return
}

func DeleteTag(session *db.Session, tagId, envId models.Id) (interface{}, e.Error) {
	if _, err := session.
		Where("id = ?", tagId).
		Where("end_id = ?", envId).
		Delete(&models.Tag{}); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return nil, nil
}

func UpdateTag(tx *db.Session, envTagId, envId models.Id, attrs models.Attrs) (envTag *models.Tag, re e.Error) {
	envTag = &models.Tag{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", envTagId).
		Where("env_id = ?", envId), &models.Tag{}, attrs); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("update tag error: %v", err))
	}
	if err := tx.Where("id = ?", envTagId).First(envTag); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query tag error: %v", err))
	}
	return
}
