// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

type ApiToken struct {
	BaseModel

	OrgId     Id `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id `json:"projectId" gorm:"size:32;default:''"`
	TplId     Id `json:"tplId" gorm:"size:32;default:''"`
	EnvId     Id `json:"envId" gorm:"size:32;default:''"`

	Token  string `json:"token" gorm:"not null;comment:Token"`
	Type   string `json:"type" gorm:"not null"`           // type:enum('api','trigger')
	Scope  string `json:"scope" gorm:"not null"`          //type:enum('org', 'project', 'template', 'env')
	Role   string `json:"role" gorm:"not null"`           // type:enum('owner','manager','operator','guest');
	Status string `json:"status" gorm:"default:'enable'"` // type:enum('enable', 'disable');

	ExpiredAt   *Time  `json:"expiredAt" gorm:"type:datetime"`
	Description string `json:"description" gorm:"type:text"`
}
