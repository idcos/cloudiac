package models

import (
	"cloudiac/portal/libs/db"
)

const (
	// 组织不允许删除

	OrgEnable  = "enable"
	OrgDisable = "disable"
)

type Organization struct {
	TimedModel

	Name        string `json:"name" gorm:"not null;unique;comment:'组织名称'"`
	Description string `json:"description" gorm:"type:text;comment:'组织描述'"`
	Status      string `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:'组织状态'"`
	CreatorId   Id     `json:"creatorId" gorm:"size:32;not null;comment:'创建人'"`
	RunnerId    string `json:"runnerId" gorm:"not null"`
}

func (Organization) TableName() string {
	return "iac_org"
}

func (o Organization) Migrate(sess *db.Session) (err error) {
	return nil
}
