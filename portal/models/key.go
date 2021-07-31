package models

import (
	"cloudiac/portal/libs/db"
)

type Key struct {
	TimedModel
	OrgId Id `json:"orgId" form:"orgId" gorm:"not null;comment:'组织ID'" example:"org-c3et0lo6n88kr92mjgq0"`

	Name      string `json:"name" gorm:"not null;comment:'密钥名称'" example:"部署密钥"`                               // 密钥名称
	Content   string `json:"-" gorm:"type:text;not null;comment:'密钥内容'" example:"xxxx"`                        // 密钥内容
	CreatorId Id     `json:"creatorId" gorm:"size:32;not null;comment:'创建人'" example:"u-c3ek0co6n88ldvq1n6ag"` //创建人ID
}

func (Key) TableName() string {
	return "iac_key"
}

func (o Key) Migrate(sess *db.Session) (err error) {
	err = o.AddUniqueIndex(sess, "unique__org__name")
	if err != nil {
		return err
	}
	return nil
}
