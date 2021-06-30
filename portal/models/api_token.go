package models

import "time"

type ApiToken struct {
	BaseModel

	OrgId     uint `json:"orgId" gorm:"not null"`
	ProjectId uint `json:"projectId" gorm:"default:'0'"`
	TplId     uint `json:"tplId" gorm:"default:'0'"`
	EnvId     uint `json:"envId" gorm:"default:'0'"`

	Token  string `json:"token" gorm:"not null;comment:'Token'"`
	Type   string `json:"type" gorm:"not null;type:enum('api','trigger')"`
	Scope  string `json:"scope" gorm:"not null;type:enum('org', 'project', 'template', 'env')"`
	Role   string `json:"role" gorm:"not null;type:enum('owner','manager','operator','guest');"`
	Status string `json:"status" gorm:"type:enum('enable', 'disable');default:'enable'"`

	ExpiredAt   *time.Time `json:"expiredAt"`
	Description string     `json:"description" gorm:"type:text"`
}
