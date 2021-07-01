package models

import (
	"cloudiac/portal/libs/db"
)

type Token struct {
	SoftDeleteModel

	UserId      Id     `json:"userId" gorm:"size:32;not null;comment:'用户ID'"`
	Token       string `json:"token" gorm:"not null;comment:'Token'"`
	Description string `json:"description" gorm:"comment:'描述'"`
	Status      string `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:'Token状态'"`
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
