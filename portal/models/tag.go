// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import "cloudiac/portal/libs/db"

type TagKey struct {
	BaseModel

	OrgId       Id     `json:"orgId" gorm:"size:32;not null"` // 组织ID
	Key         string `json:"key" gorm:"not null"`           // tag key
	Description string `json:"description"`                   // key 描述
}

func (TagKey) TableName() string {
	return "iac_tag_key"
}

func (o TagKey) Migrate(sess *db.Session) (err error) {
	if err := o.AddUniqueIndex(sess, "unique__org__key",
		"org_id", "`key`"); err != nil {
		return err
	}
	return nil
}

type TagValue struct {
	BaseModel

	OrgId Id     `json:"orgId" gorm:"size:32;not null"` // 组织ID
	KeyId Id     `json:"keyId" gorm:"size:32;not null"`
	Value string `json:"value" gorm:"not null"`
}

func (o TagValue) Migrate(sess *db.Session) (err error) {
	if err := o.AddUniqueIndex(sess, "unique__org__value__key",
		"org_id", "key_id", "value"); err != nil {
		return err
	}
	return nil
}

func (TagValue) TableName() string {
	return "iac_tag_value"
}

type TagRel struct {
	AbstractModel

	OrgId      Id     `json:"orgId" gorm:"size:32;not null"` // 组织ID
	TagKeyId   Id     `json:"tagKeyId" gorm:"size:32;not null"`
	TagValueId Id     `json:"tagValueId" gorm:"size:32;not null"`
	ObjectId   Id     `json:"objectId" gorm:"size:32;not null"`
	ObjectType string `json:"objectType" gorm:"not null;type:enum('env')"`
	Source     string `json:"source" gorm:"not null"`
}

func (TagRel) TableName() string {
	return "iac_tag_rel"
}

func (o TagRel) Migrate(sess *db.Session) (err error) {
	if err := o.AddUniqueIndex(sess, "unique__org__value__key__object",
		"org_id", "tag_key_id", "object_id", "object_type"); err != nil {
		return err
	}
	return nil
}

type TagValueWithSource struct {
	Key     string `json:"key"`
	KeyId   Id     `json:"keyId"`
	Value   string `json:"value"`
	ValueId Id     `json:"valueId"`
	Source  string `json:"source"`
}
