// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/portal/libs/db"
)

type ResourceAccount struct {
	TimedModel

	OrgId       Id     `json:"-" gorm:"size:32;not null;comment:组织ID"`
	Name        string `json:"name" gorm:"size:32;not null;comment:资源账号名称"`
	Description string `json:"description" gorm:"size:255;comment:资源账号描述"`
	Params      JSON   `json:"params" gorm:"type:text;null;comment:账号变量"`
	Status      string `json:"status" gorm:"default:'enable';comment:资源账号状态"` // type:enum('enable','disable');
}

func (ResourceAccount) TableName() string {
	return "iac_resource_account"
}

func (r ResourceAccount) Migrate(sess *db.Session) (err error) {
	err = r.AddUniqueIndex(sess, "unique__org_id__name", "org_id", "name")
	if err != nil {
		return err
	}

	return nil
}

type CtResourceMap struct {
	BaseModel

	ResourceAccountId Id     `json:"resourceAccountId" gorm:"size:32;not null;comment:资源账号ID"`
	CtServiceId       string `json:"ctServiceId" gorm:"size:64;not null;comment:Runner Service ID"`
}

func (CtResourceMap) TableName() string {
	return "iac_ct_resource_map"
}

func (c CtResourceMap) Migrate(sess *db.Session) (err error) {
	err = c.AddUniqueIndex(sess, "unique__resource_account_id__ct_service_id", "resource_account_id", "ct_service_id")
	if err != nil {
		return err
	}

	return nil
}
