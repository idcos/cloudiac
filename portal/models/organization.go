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
	Guid        string `json:"guid" gorm:"size:32;not null;unique;comment:'组织GUID'"`
	Description string `json:"description" gorm:"type:text;comment:'组织描述'"`
	Status      string `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:'组织状态'"`
	CreatorId   uint   `json:"creatorId" gorm:"not null;comment:'创建人'"`
	// TODO 以下两个字段需要删除，只保留 runnerId，地址和端口在执行任务的时候实时查询
	DefaultRunnerAddr string `json:"defaultRunnerAddr" gorm:"not null;comment:'默认runner地址'"`
	DefaultRunnerPort uint   `json:"defaultRunnerPort" gorm:"not null;comment:'默认runner端口'"`
	RunnerId          string `json:"runnerId" gorm:"not null"`
}

func (Organization) TableName() string {
	return "iac_org"
}

func (o Organization) Migrate(sess *db.Session) (err error) {
	return nil
}
