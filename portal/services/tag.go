// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"fmt"
)

func SearchTag(dbSess *db.Session, orgId, objectId models.Id, objectType string) *db.Session {
	tr := models.TagRel{}.TableName()
	query := dbSess.Table(fmt.Sprintf("%s as tr", tr)).
		Where("tr.org_id = ?", orgId).
		Where("tr.object_id = ?", objectId).
		Where("tr.object_type = ?", objectType)

	query = query.
		Joins(fmt.Sprintf("left join %s as tv on tr.tag_value_id = tv.id",
			models.TagValue{}.TableName()))

	query = query.
		Joins(fmt.Sprintf("left join %s as tk on tr.tag_key_id = tk.id",
			models.TagKey{}.TableName()))

	return query.
		LazySelectAppend("tv.id as value_id", "tk.id as key_id", "tv.value", "tk.key", "tr.source").
		Order("tr.source, tk.key")
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
	tagKey := make([]models.TagKey, 0)
	if err := session.
		Where("`key` = ?", key).
		Where("org_id = ?", orgId).
		Find(&tagKey); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return tagKey, nil
}

func FindTagValueByName(session *db.Session, value string, keyId, orgId models.Id) ([]models.TagValue, e.Error) {
	tagValue := make([]models.TagValue, 0)
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
		tagKey.Id = models.NewId("tk-")
	}

	if err := models.Create(tx, &tagKey); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return &tagKey, nil
}

func CreateTagValue(tx *db.Session, tagValue models.TagValue) (*models.TagValue, e.Error) {
	if tagValue.Id == "" {
		tagValue.Id = models.NewId("tv-")
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

// 查询 tagKey，不存在则创建并返回
func FindOrCreateTagKeys(tx *db.Session, orgId models.Id, keys []string) (map[string]*models.TagKey, e.Error) {
	dbTagKeys := make([]*models.TagKey, 0)
	err := tx.Model(&models.TagKey{}).
		Where("org_id = ? AND `key` IN (?)", orgId, keys).
		Find(&dbTagKeys)
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	newTagKeys := make([]*models.TagKey, 0)
	for _, k := range keys {
		exists := false
		for _, tk := range dbTagKeys {
			if k == tk.Key {
				exists = true
				break
			}
		}
		if !exists {
			newTagKeys = append(newTagKeys, &models.TagKey{
				OrgId: orgId,
				Key:   k,
			})
		}
	}

	if len(newTagKeys) > 0 {
		if err := models.CreateBatch(tx, newTagKeys); err != nil {
			return nil, e.AutoNew(err, e.DBError)
		}
	}

	rs := make(map[string]*models.TagKey)
	for _, tk := range append(dbTagKeys, newTagKeys...) {
		rs[tk.Key] = tk
	}
	return rs, nil
}

// 更新或创建 TagValues
// tags 入参的 map key 为 tagKeyId
// 返回值的 map key 也为 tagKeyId
func UpsertTagValues(tx *db.Session, orgId models.Id, tags map[models.Id]string) (map[models.Id]*models.TagValue, e.Error) {

	keyIds := make([]models.Id, 0, len(tags))
	for k := range tags {
		keyIds = append(keyIds, k)
	}

	dbTagVals := make([]*models.TagValue, 0)
	err := tx.Model(&models.TagValue{}).
		Where("org_id = ? AND `key_id` IN (?)", orgId, keyIds).
		Find(&dbTagVals)
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	// 更新 tag value
	for _, tv := range dbTagVals {
		newVal := tags[tv.KeyId]
		if tv.Value != newVal {
			tv.Value = newVal
			if err := models.Save(tx, tv); err != nil {
				return nil, e.AutoNew(err, e.DBError)
			}
		}
	}

	newTagVals := make([]*models.TagValue, 0)
	for kid := range tags {
		exists := false
		for _, tv := range dbTagVals {
			if kid == tv.KeyId {
				exists = true
				break
			}
		}
		if !exists {
			newTagVals = append(newTagVals, &models.TagValue{
				OrgId: orgId,
				KeyId: kid,
				Value: tags[kid],
			})
		}
	}

	// 插入 tag value
	if len(newTagVals) > 0 {
		if err := models.CreateBatch(tx, newTagVals); err != nil {
			return nil, e.AutoNew(err, e.DBError)
		}
	}

	rs := make(map[models.Id]*models.TagValue)
	for _, tv := range append(dbTagVals, newTagVals...) {
		rs[tv.KeyId] = tv
	}
	return rs, nil
}

func UpInsertTags(tx *db.Session, orgId models.Id, tags map[string]string) (map[string]*models.TagValue, e.Error) {
	keys := []string{}
	for k := range tags {
		keys = append(keys, k)
	}

	tKeys, er := FindOrCreateTagKeys(tx, orgId, keys)
	if er != nil {
		return nil, er
	}

	tagKeyIdMap := make(map[models.Id]string, 0)
	for k, v := range tKeys {
		tagKeyIdMap[v.Id] = tags[k]
	}

	tVals, er := UpsertTagValues(tx, orgId, tagKeyIdMap)
	if er != nil {
		return nil, er
	}

	rs := make(map[string]*models.TagValue)
	for kid, v := range tVals {
		k := tagKeyIdMap[kid]
		rs[k] = v
	}
	return rs, nil
}

// 全量更新对象的 tags: 删除旧的同 source tags，将新的 tags 绑定到对象
func UpdateObjectTags(tx *db.Session, orgId, objId models.Id, objType, source string, tags map[string]string) ([]*models.TagRel, e.Error) {
	_, err := tx.
		Where("org_id = ? AND object_id = ? AND object_type = ? AND source = ?", orgId, objId, objType, source).
		Delete(&models.TagRel{})
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return AddTagsToObject(tx, orgId, objId, objType, source, tags)
}

// 添加 tags 到对象上，如果对象已有相同的 tag key 则更新 tag value
func AddTagsToObject(tx *db.Session, orgId, objId models.Id, objType, source string, tags map[string]string) ([]*models.TagRel, e.Error) {
	dbTags, er := UpInsertTags(tx, orgId, tags)
	if er != nil {
		return nil, er
	}

	// 检查对象 tag 数量
	n, err := tx.Model(&models.TagRel{}).
		Where("org_id = ? AND object_id = ? AND object_type = ?", orgId, objId, objType).
		Count()
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	if n > consts.ObjectMaxTagNum {
		return nil, e.New(e.ObjectTagNumLimited)
	}

	bs := utils.NewBatchSQL(1024, "REPLACE INTO", models.TagRel{}.TableName(),
		"org_id", "tag_key_id", "tag_value_id", "object_id", "object_type", "source")

	tagKeyIds := make([]models.Id, 0)
	for _, tv := range dbTags {
		bs.MustAddRow(orgId, tv.KeyId, tv.Id, objId, objType, source)
		tagKeyIds = append(tagKeyIds, tv.KeyId)
	}

	for bs.HasNext() {
		sql, args := bs.Next()
		if _, err := tx.Exec(sql, args...); err != nil {
			return nil, e.AutoNew(err, e.DBError)
		}
	}

	rels := make([]*models.TagRel, 0, len(tagKeyIds))
	if err := tx.Model(&models.TagRel{}).
		Where("org_id = ? AND object_id = ? AND object_type = ?", orgId, objId, objType).
		Where("tag_key_id IN (?)", tagKeyIds).
		Find(&rels); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return rels, nil
}

func FindObjectTags(db *db.Session, orgId, objId models.Id, objType string) ([]*models.TagValueWithSource, e.Error) {
	tags := make([]*models.TagValueWithSource, 0)
	if err := SearchTag(db, orgId, objId, objType).Find(&tags); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return tags, nil
}

func FindObjectTagMap(db *db.Session, orgId, objId models.Id, objType string) (rs map[string]string, er e.Error) {
	tags, er := FindObjectTags(db, orgId, objId, objType)
	if er != nil {
		return nil, er
	}

	rs = make(map[string]string)
	for _, t := range tags {
		rs[t.Key] = t.Value
	}
	return rs, nil
}
