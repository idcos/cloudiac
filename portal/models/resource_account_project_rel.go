// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

type ResourceAccountProjectRel struct {
	AbstractModel
	ProjectId         Id `json:"projectId" gorm:"size:32;not null;comment:项目ID"`
	ResourceAccountId Id `json:"resourceAccountId" gorm:"size:32;not null;comment:资源账号ID"`
}

func (ResourceAccountProjectRel) TableName() string {
	return "iac_resource_account_project_rel"
}