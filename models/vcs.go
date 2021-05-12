package models

import "cloudiac/libs/db"

type Vcs struct {
	BaseModel
	OrgId   uint `json:"orgId" gorm:"not null;" comment:"组织id"`
	Name    string `json:"name" gorm:"not null;comment:'vcs名称'"`
	VcsType string `json:"vcsType" gorm:"not null;comment:'vcs代码库类型'"`
	Status  string `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:'vcs状态'"`
	Address string `json:"address" gorm:"not null;comment:'vcs代码库地址'"`
	Token   string `json:"vcsToken" gorm:"not null; comment:'代码库的token值'"`
}

func (Vcs) TableName() string {
	return "iac_vcs"
}

func (o Vcs) Migrate(sess *db.Session) (err error) {
	return nil
}
