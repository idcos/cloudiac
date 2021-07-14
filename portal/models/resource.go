package models

import "database/sql/driver"

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
	TaskId    Id `json:"taskId" gorm:"size:32;not null"`

	Provider string   `json:"provider" gorm:"not null"`
	Module   string   `json:"module,omitempty" gorm:"not null;default:''"`
	Address  string   `json:"address" gorm:"not null"`
	Type     string   `json:"type" gorm:"not null"`
	Name     string   `json:"name" gorm:"not null"`
	Index    int      `json:"index" gorm:"not null"`
	Attrs    ResAttrs `json:"attrs" gorm:"type:json"`
}

func (Resource) TableName() string {
	return "iac_resource"
}
