package models

import (
	"cloudiac/libs/db"
)

type User struct {
	SoftDeleteModel

	Name        string `json:"name" gorm:"size:32;not null;comment:'姓名'"`
	Email       string `json:"email" gorm:"size:64;not null;comment:'邮箱'"`
	Password    string `json:"-" gorm:"not null;comment:'密码'"`
	Phone       string `json:"phone" gorm:"size:16;comment:'电话'"`
	IsAdmin     bool   `json:"isAdmin" gorm:"default:false;comment:'是否平台管理员'"`
	Status      string `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:'用户状态'"`
	NewbieGuide JSON   `json:"newbieGuide" gorm:"type:json;null;comment:'新手引导状态'"`
}

func (User) TableName() string {
	return "iac_user"
}

func (u User) Migrate(sess *db.Session) (err error) {
	err = u.AddUniqueIndex(sess, "unique__email", "email")
	if err != nil {
		return err
	}

	return nil
}

type UserOrgMap struct {
	BaseModel

	OrgId  uint   `json:"orgId" gorm:"not null;comment:'组织ID'"`
	UserId uint   `json:"userId" gorm:"not null;comment:'用户ID'"`
	Role   string `json:"role" gorm:"type:enum('owner','member');default:'member';comment:'角色'"`
}

func (UserOrgMap) TableName() string {
	return "iac_user_org_map"
}

func (m UserOrgMap) Migrate(sess *db.Session) (err error) {
	err = m.AddUniqueIndex(sess, "unique__org_id__user_id", "org_id", "user_id")
	if err != nil {
		return err
	}

	return nil
}
