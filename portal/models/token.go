package models

import (
	"cloudiac/portal/libs/db"
	"time"
)

type Token struct {
	SoftDeleteModel

	Key         string     `json:"key" form:"key" gorm:"not null"`
	Type        string     `json:"type" form:"type" gorm:"not null"`
	OrgId       Id         `json:"orgId" form:"orgId" gorm:"not null"`
	Role        string     `json:"role" form:"role" gorm:"not null"`
	Status      string     `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:'Token状态'"`
	ExpiredAt   *time.Time `json:"ExpiredAt" form:"ExpiredAt" gorm:"not null"`
	Description string     `json:"description" gorm:"comment:'描述'"`
}

func (Token) TableName() string {
	return "iac_token"
}

func (o Token) Migrate(sess *db.Session) (err error) {
	err = o.AddUniqueIndex(sess, "unique__token", "token")
	if err != nil {
		return err
	}
	return nil
}

type LoginResp struct {
	//UserInfo *models.User
	Token string `json:"token" example:"eyJhbGciO..."` // 登陆令牌
}
