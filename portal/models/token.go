// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/portal/libs/db"
)

type Token struct {
	TimedModel

	Name        string `json:"name" form:"name" gorm:"unique;comment:Token名称"`
	Key         string `json:"key" form:"key" gorm:"not null"`
	Type        string `json:"type" form:"type" gorm:"not null"`
	OrgId       Id     `json:"orgId" form:"orgId" gorm:"not null"`
	Role        string `json:"role" form:"role" gorm:"not null"`
	Status      string `json:"status" gorm:"default:'enable';comment:Token状态"` // type:enum('enable','disable');
	ExpiredAt   *Time  `json:"expiredAt" form:"expiredAt" gorm:""`
	Description string `json:"description" gorm:"comment:描述"`
	CreatorId   Id     `json:"creatorId" gorm:"size:32;not null;comment:创建人" example:"u-c3ek0co6n88ldvq1n6ag"` //创建人ID

	// 触发器需要的字段
	EnvId  Id     `json:"envId" form:"envId"  gorm:"not null"`
	Action string `json:"action" form:"action" gorm:"default:'plan'"` // type:enum('apply','plan','destroy');
}

func (Token) TableName() string {
	return "iac_token"
}

func (Token) NewId() Id {
	return NewId("t")
}

func (o Token) Migrate(sess *db.Session) (err error) {
	err = o.AddUniqueIndex(sess, "unique__key", "`key`")
	if err != nil {
		return err
	}

	if _, err = sess.Exec("update iac_token set iac_token.`name` = substr(`key`, 1 ,8) where name IS NULL;"); err != nil {
		return err
	}
	return nil
}
