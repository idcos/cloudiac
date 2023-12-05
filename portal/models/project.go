// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import "cloudiac/portal/libs/db"

type Project struct {
	SoftDeleteModel

	OrgId       Id     `json:"orgId" gorm:"size:32;not null"`             //组织ID
	Name        string `json:"name" form:"name" gorm:"not null;"`         //组织名称
	Description string `json:"description" gorm:"type:text"`              //组织详情
	CreatorId   Id     `json:"creatorId" form:"creatorId" `               //用户id
	Status      string `json:"status" gorm:"default:'enable';comment:状态"` // type:enum('enable','disable');

	IsDemo bool `json:"isDemo"`
}

func (Project) TableName() string {
	return "iac_project"
}

func (p *Project) Migrate(sess *db.Session) (err error) {
	if err := p.AddUniqueIndex(sess,
		"unique__org__project__name", "org_id", "name"); err != nil {
		return err
	}
	return nil
}
