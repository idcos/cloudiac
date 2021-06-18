package models

import "cloudiac/libs/db"

type MetaTemplate struct {
	TimedModel
	Name        string `json:"name" gorm:"size:32;not null;comment:'模版名称'"`
	Description string `json:"description" gorm:"size:255;comment:'描述'"`
	RepoId      string `json:"repoId" gorm:"size:128;comment:'仓库ID'"`
	RepoAddr    string `json:"repoAddr" gorm:"size:255;comment:'仓库地址'"`
	RepoBranch  string `json:"repoBranch" gorm:"size:64;default:'master';comment:'仓库分支'"`
	SaveState   bool   `json:"saveState" gorm:"comment:'是否保存状态'"`
	Vars        JSON   `json:"vars" gorm:"type:json;null;comment:'变量'"`
	Varfile     string `json:"varfile" gorm:"size:128;default:'';comment:'变量文件'"`
	Timeout     int64  `json:"timeout" gorm:"default:300;comment:'超时时长'"`
	VcsId       uint   `json:"vcsId" gorm:"not null;"`
	Playbook    string `json:"playbook" form:"playbook" `
}

func (MetaTemplate) TableName() string {
	return "iac_meta_template"
}

func (t *MetaTemplate) Migrate(sess *db.Session) (err error) {
	if err := sess.DB().ModifyColumn("repo_id",
		"VARCHAR(128) NOT NULL COMMENT '仓库 Id 或 Path'").Error; err != nil {
		return err
	}

	if err := sess.DB().ModifyColumn("repo_addr",
		"VARCHAR(255) NOT NULL COMMENT '仓库地址'").Error; err != nil {
		return err
	}

	return nil
}
