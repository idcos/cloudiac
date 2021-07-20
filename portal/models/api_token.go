package models

import (
	"cloudiac/utils"
)

type ApiToken struct {
	BaseModel

	OrgId     Id `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id `json:"projectId" gorm:"size:32;default:'0'"`
	TplId     Id `json:"tplId" gorm:"size:32;default:'0'"`
	EnvId     Id `json:"envId" gorm:"size:32;default:'0'"`

	Token  string `json:"token" gorm:"not null;comment:'Token'"`
	Type   string `json:"type" gorm:"not null;type:enum('api','trigger')"`
	Scope  string `json:"scope" gorm:"not null;type:enum('org', 'project', 'template', 'env')"`
	Role   string `json:"role" gorm:"not null;type:enum('owner','manager','operator','guest');"`
	Status string `json:"status" gorm:"type:enum('enable', 'disable');default:'enable'"`

	ExpiredAt   *utils.JSONTime `json:"expiredAt"`
	Description string          `json:"description" gorm:"type:text"`
}
