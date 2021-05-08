package models

import "cloudiac/libs/db"

type Vcs struct {
	BaseModel
	Name    string `json:"name" gorm:"not null;comment:'vcs名称'"`
	VcsType string `json:"vcs_type" gorm:"not null;comment:'vcs代码库类型'"`
	Status  string `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:'vcs状态'"`
	Address string `json:"address" gorm:"not null;comment:'vcs代码库地址'"`
	Token   string `json:"token" gorm:"not null; comment:'代码库的token值'"`
}

func (Vcs) TableName() string {
	return "iac_vcs"
}

func (o Vcs) Migrate(sess *db.Session) (err error) {
	return nil
}
