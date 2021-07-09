package models

import "cloudiac/portal/libs/db"

type Template struct {
	SoftDeleteModel

	Name         string `json:"name" gorm:"not null;comment:'模版名称'"`
	TplType      string `json:"tplType" gorm:"not null;comment:'云模板类型(aliyun，VMware等)'"`
	OrgId        Id     `json:"orgId" gorm:"size:32;not null"`
	// TODO 模板己改为在组织层管理，是否还需要 projectId 字段？
	ProjectId    Id     `json:"projectId" gorm:"size:32;not null"`
	Description  string `json:"description" gorm:"type:text"`
	VcsId        Id     `json:"vcsId" gorm:"size:32;not null"`
	RepoId       string `json:"repoId" gorm:"not null"`
	RepoAddr     string `json:"repoAddr" gorm:"not null"`  // RepoAddr 可以为相对路径，以支持修改 vcs 的地址
	RepoToken    string `json:"repoToken" gorm:"size:128"` // RepoToken 若为空则使用 vcs 的 token
	RepoRevision string `json:"repoRevision" gorm:"size:64;default:'master'"`

	// 模板的默认 runner 可以为空
	RunnerId  string `json:"runnerId" gorm:"size:32;comment:'默认 runnerId'"`
	Status    string `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:'状态'"`
	CreatorId Id     `json:"creatorId" gorm:"size:32;not null;comment:'创建人'"`

	Workdir string `json:"workdir" gorm:"default:''"` // 是基于项目根目录的相对路径, 默认为项目根目录

	// 要执行的 ansible playbook 文件(相对于 workdir 的路径)
	Playbook string `json:"playbook" gorm:"default:''"`

	// 页面上定义的 terraform variable 会覆盖 tfvars 文件里的值(相对于 workdir 的路径)
	TfVarsFile string `json:"tfVarsFile" gorm:"default:''"`

	// 页面上定义的 ansible variable 覆盖 playVars 文件里的值，通过传参控制，
	// ansible-playbook --extra-vars 参数可以指定多次，后传入的同名变量覆盖先传入的
	PlayVarsFile string `json:"playVarsFile" gorm:"default:''"`
}

func (Template) TableName() string {
	return "iac_template"
}

func (t *Template) Migrate(sess *db.Session) (err error) {
	if err = sess.RemoveIndex("iac_template", "unique__project__tpl__name"); err != nil {
		return err
	}
	if err = t.AddUniqueIndex(sess, "unique__org__tpl__name", "org_id", "name"); err != nil {
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
