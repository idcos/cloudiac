package models

import "cloudiac/portal/libs/db"

type Template struct {
	SoftDeleteModel

	Name         string `json:"name" gorm:"not null;comment:'模版名称'"`
	Guid         string `json:"guid" gorm:"size:32;not null;unique"`
	TplType      string `json:"tplType" gorm:"not null;comment:'云模板类型(aliyun，VMware等)'"`
	OrgId        uint   `json:"orgId" gorm:"not null"`
	ProjectId    uint   `json:"projectId" gorm:"not null"`
	Description  string `json:"description" gorm:"type:text"`
	VcsId        uint   `json:"vcsId" gorm:"not null"`
	RepoId       string `json:"repoId" gorm:"not null"`
	RepoAddr     string `json:"repoAddr" gorm:"not null"`
	RepoRevision string `json:"repoRevision" gorm:"size:64;default:'master'"`
	Workdir      string `json:"workdir" gorm:"default:''"`  // 基于项目根目录的相对路径
	Playbook     string `json:"playbook" gorm:"default:''"` // 基于项目根目录的相对路径
	Status       string `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:'状态'"`
	CreatorId    uint   `json:"creatorId" gorm:"not null;comment:'创建人'"`
	RunnerId     string `json:"runnerId" gorm:"not null;comment:'默认 runnerId'"`
}

func (Template) TableName() string {
	return "iac_template"
}

func (t *Template) Migrate(sess *db.Session) (err error) {
	if err := t.AddUniqueIndex(sess, "unique__project__tpl__name", "project_id", "name"); err != nil {
		return err
	}
	return nil
}

// TODO 改用统一的 ApiToken 表
type TemplateAccessToken struct {
	TimedModel

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
