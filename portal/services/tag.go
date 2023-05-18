// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

func SearchTag(dbSess *db.Session, orgId, objectId models.Id, objectType string) *db.Session {
	tr := models.TagRel{}.TableName()
	query := dbSess.Table(fmt.Sprintf("%s as tr",tr)).
		Where("tr.org_id = ?", orgId).
		Where("tr.object_id = ?", objectId).
		Where("tr.object_type = ?", objectType)

	query = query.
		Joins(fmt.Sprintf("left join %s as tv on tr.tag_value_id = tv.id",
			models.TagValue{}.TableName()))

	query = query.
		Joins(fmt.Sprintf("left join %s as tk on tr.tag_key_id = tk.id",
			models.TagKey{}.TableName()))

	return query.LazySelectAppend("tv.id as value_id", "tk.id as key_id", "tv.value", "tk.key").Order("tk.id")
}

func DeleteTagRel(tx *db.Session, keyId, valueId, orgId, objectId models.Id, objectType string) e.Error {
	if _, err := tx.
		Where("tag_key_id = ?", keyId).
		Where("tag_value_id = ?", valueId).
		Where("org_id = ?", orgId).
		Where("object_id = ?", objectId).
		Where("object_type = ?", objectType).
		Delete(&models.TagRel{}); err != nil {
		return e.New(e.DBError, err)
	}

	return nil
}


func FindTagKeyByName(session *db.Session, key string, orgId models.Id) ([]models.TagKey, e.Error) {
	tagKey := make([]models.TagKey,0)
	if err := session.
		Where("`key` = ?", key).
		Where("org_id = ?", orgId).
		Find(&tagKey); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return tagKey, nil
}

func FindTagValueByName(session *db.Session, value string, keyId,orgId models.Id) ([]models.TagValue, e.Error) {
	tagValue := make([]models.TagValue,0)
	if err := session.
		Where("value = ?", value).
		Where("org_id = ?", orgId).
		Where("key_id = ?", keyId).
		Find(&tagValue); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return tagValue, nil
}

func FindTagRelById(session *db.Session, keyId, valueId, orgId, objectId models.Id, objectType string) ([]models.TagRel, e.Error) {
	tagRel := make([]models.TagRel, 0)
	if err := session.
		Where("tag_key_id = ?", keyId).
		Where("tag_value_id = ?", valueId).
		Where("object_id = ?", objectId).
		Where("object_type = ?", objectType).
		Where("org_id = ?", orgId).
		Find(&tagRel); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return tagRel, nil
}

func CreateTagKey(tx *db.Session, tagKey models.TagKey) (*models.TagKey, e.Error) {
	if tagKey.Id == "" {
		tagKey.Id = models.NewId("tagk-")
	}

	if err := models.Create(tx, &tagKey); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return &tagKey, nil
}

func CreateTagValue(tx *db.Session, tagValue models.TagValue) (*models.TagValue, e.Error) {
	if tagValue.Id == "" {
		tagValue.Id = models.NewId("tagv-")
	}

	if err := models.Create(tx, &tagValue); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return &tagValue, nil
}

func CreateTagRel(tx *db.Session, tagRel models.TagRel) (*models.TagRel, e.Error) {
	if err := models.Create(tx, &tagRel); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return &tagRel, nil
}
