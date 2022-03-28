// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

import (
	"database/sql/driver"
)

type ResAttrs map[string]interface{}

func (v ResAttrs) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *ResAttrs) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

type Resource struct {
	BaseModel

	OrgId     Id `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id `json:"projectId" gorm:"size:32;not null"`
	EnvId     Id `json:"envId" gorm:"size:32;not null"`
	TaskId    Id `json:"taskId" gorm:"index;size:32;not null"`

	Provider      string   `json:"provider" gorm:"not null"`
	Module        string   `json:"module,omitempty" gorm:"not null;default:''"`
	Address       string   `json:"address" gorm:"not null"`
	Type          string   `json:"type" gorm:"not null"`
	Name          string   `json:"name" gorm:"not null"`
	Index         string   `json:"index" gorm:"not null;default:''"`
	Attrs         ResAttrs `json:"attrs,omitempty" gorm:"type:json"`
	SensitiveKeys StrSlice `json:"sensitiveKeys,omitempty" gorm:"type:json"`
	AppliedAt     Time     `json:"appliedAt" gorm:"type:datetime;column:applied_at;default:null"`
	ResId         Id       `json:"resId" gorm:"not null"`
}

func (Resource) TableName() string {
	return "iac_resource"
}
