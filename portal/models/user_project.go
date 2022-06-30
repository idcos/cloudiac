// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/portal/libs/db"
)

type UserProject struct {
	AutoUintIdModel

	UserId     Id     `json:"userId" gorm:"size:32;not null;comment:用户ID"`
	ProjectId  Id     `json:"projectId" gorm:"size:32;not null"`
	Role       string `json:"role" gorm:"type:enum('manager','approver','operator','guest');default:'operator';comment:角色"`
	IsFromLdap bool   `json:"isFromLdap" gorm:"default:false"` // 权限是否来自ldap
}

func (UserProject) TableName() string {
	return "iac_user_project"
}

func (u UserProject) Migrate(sess *db.Session) error {
	return u.AddUniqueIndex(sess, "unique__user__project", "user_id", "project_id")
}
