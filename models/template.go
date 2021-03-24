package models

import "cloudiac/libs/db"

type Template struct {
	SoftDeleteModel

	Name        string `json:"name" gorm:"size:32;not null;comment:'模版名称'"`
	Guid        string `json:"guid" gorm:"size:32;not null;comment:'模板GUID'"`
	OrgId       string `json:"org_id" gorm:"size:32;not null;comment:'组织ID'"`
	Description string `json:"description" gorm:"size:255;comment:'描述'"`
	RepoId      int    `json:"repo_id" gorm:"size:32;comment:'仓库ID'"`
	RepoAddr    string `json:"repo_addr" gorm:"size:128;default:'';comment:'仓库地址'"`
	RepoBranch  string `json:"repo_branch" gorm:"size:64;default:'master';comment:'仓库分支'"`
	SaveState   bool   `json:"save_state" gorm:"defalut:false;comment:'是否保存状态'"`
	Vars        JSON   `json:"vars" gorm:"type:json;null;comment:'变量'"`
	Varfile     string `json:"varfile" gorm:"size:128;default:'';comment:'变量文件'"`
	Extra       string `json:"extra" gorm:"size:128;default:'';comment:'附加信息'"`
	Timeout     int    `json:"timeout" gorm:"size:32;comment:'超时时长'"`
	Creator     uint   `json:"creator" grom:"not null;comment:'创建人'"`
}

func (Template) TableName() string {
	return "iac_template"
}

func (o Template) Migrate(sess *db.Session) (err error) {
	err = o.AddUniqueIndex(sess, "unique__guid", "guid")
	if err != nil {
		return err
	}

	return nil
}
