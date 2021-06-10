package models

import "cloudiac/libs/db"

type Template struct {
	SoftDeleteModel

	Name                   string `json:"name" gorm:"size:32;not null;comment:'模版名称'"`
	Guid                   string `json:"guid" gorm:"size:32;not null;comment:'模板GUID'"`
	OrgId                  uint   `json:"orgId" gorm:"size:32;not null;comment:'组织ID'"`
	Description            string `json:"description" gorm:"size:255;comment:'描述'"`
	RepoId                 string `json:"repoId"` // see Migrate()
	RepoAddr               string `json:"repoAddr"` // see Migrate()
	RepoBranch             string `json:"repoBranch" gorm:"size:64;default:'master';comment:'仓库分支'"`
	SaveState              bool   `json:"saveState" gorm:"default:false;comment:'是否保存状态'"`
	Vars                   JSON   `json:"vars" gorm:"type:json;null;comment:'变量'"`
	Varfile                string `json:"varfile" gorm:"size:128;default:'';comment:'变量文件'"`
	Extra                  string `json:"extra" gorm:"size:128;default:'';comment:'附加信息'"`
	Timeout                int64  `json:"timeout" gorm:"default:300;comment:'超时时长'"`
	Status                 string `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:'状态'"`
	Creator                uint   `json:"creator" gorm:"not null;comment:'创建人'"`
	DefaultRunnerAddr      string `json:"defaultRunnerAddr" gorm:"not null;comment:'默认runner地址'"`
	DefaultRunnerPort      uint   `json:"defaultRunnerPort" gorm:"not null;comment:'默认runner端口'"`
	DefaultRunnerServiceId string `json:"defaultRunnerServiceId" gorm:"not null;comment:'默认runner-consul-serviceId'"`
	TplType                string `json:"tplType" gorm:"not null;comment:'云模板类型(aliyun，VMware等)'"`
	VcsId                  uint   `json:"vcsId" gorm:"not null;"`
	Playbook               string `json:"playbook" form:"playbook" `
}

func (Template) TableName() string {
	return "iac_template"
}

func (o Template) Migrate(sess *db.Session) (err error) {
	err = o.AddUniqueIndex(sess, "unique__guid", "guid")
	if err != nil {
		return err
	}

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

type TemplateAccessToken struct {
	BaseModel

	TplGuid     string `json:"tplGuid" form:"tplGuid" gorm:"not null"`
	AccessToken string `json:"accessToken" form:"accessToken" gorm:"not null"`
	Action      string `json:"action" form:"action"  gorm:"type:enum('plan','apply','compliance');default:'plan'"`
}

func (TemplateAccessToken) TableName() string {
	return "iac_template_access_token"
}

func (o TemplateAccessToken) Migrate(sess *db.Session) (err error) {
	err = o.AddUniqueIndex(sess, "unique__guid", "access_token")
	if err != nil {
		return err
	}

	return nil
}
