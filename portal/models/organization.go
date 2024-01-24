// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

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

	Name        string `json:"name" gorm:"not null;unique;comment:组织名称" example:"研发部"`                              // 组织名称
	Description Text   `json:"description" gorm:"type:text;comment:组织描述" example:"示例公司 IaC 研发部"`                    // 组织描述
	Status      string `json:"status" gorm:"default:'enable';comment:组织状态" example:"enable" enums:"enable,disable"` // 组织状态
	CreatorId   Id     `json:"creatorId" gorm:"size:32;not null;comment:创建人" example:"u-c3ek0co6n88ldvq1n6ag"`      //创建人ID
	RunnerId    string `json:"runnerId" gorm:"" example:"runner-01"`                                                // 组织默认部署通道

	IsDemo bool `json:"isDemo" gorm:"default:false"` // 是否演示组织
}

func (Organization) TableName() string {
	return "iac_org"
}

func (o Organization) Migrate(sess *db.Session) (err error) {
	return nil
}
