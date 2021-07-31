// Copyright 2021 CloudJ Company Limited. All rights reserved.

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

	Name        string `json:"name" gorm:"not null;unique;comment:'组织名称'" example:"研发部"`                                                            // 组织名称
	Description string `json:"description" gorm:"type:text;comment:'组织描述'" example:"示例公司 IaC 研发部"`                                                  // 组织描述
	Status      string `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:'组织状态'" example:"enable" enums:"enable,disable"` // 组织状态
	CreatorId   Id     `json:"creatorId" gorm:"size:32;not null;comment:'创建人'" example:"u-c3ek0co6n88ldvq1n6ag"`                                    //创建人ID
	RunnerId    string `json:"runnerId" gorm:"not null" example:"iac-porta;portal-01"`                                                              // 组织默认部署通道

	IsDemo bool `json:"isDemo,omitempty" gorm:"default:'0'"` // 是否演示组织
}

func (Organization) TableName() string {
	return "iac_org"
}

func (o Organization) Migrate(sess *db.Session) (err error) {
	return nil
}

type OrgDetailResp struct {
	Organization
	Creator string `json:"creator" example:"研发部负责人"` // 创建人名称
}
