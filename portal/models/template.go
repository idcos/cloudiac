// Copyright 2021 CloudJ Company Limited. All rights reserved.

package models

import (
	"cloudiac/portal/libs/db"
	"github.com/lib/pq"
)

type Template struct {
	SoftDeleteModel

	Name        string `json:"name" gorm:"not null;comment:模板名称" example:"yunji_example"`
	TplType     string `json:"tplType" gorm:"not null;comment:云模板类型(aliyun，VMware等)" example:"aliyun"`
	OrgId       Id     `json:"orgId" gorm:"size:32;not null" example:"a1f79e8a-744d-4ea5-8d97-7e4b7b422a6c"`
	Description string `json:"description" gorm:"type:text" example:"云霁阿里云模板"`

	// 如果创建模板时用户直接填写完整 RepoAddr 则 vcsId 为空值，
	// 此时创建任务直接使用 RepoRevision 做为 commit id，不再实时获取
	VcsId Id `json:"vcsId" gorm:"size:32;not null" example:"a1f79e8a-744d-4ea5-8d97-7e4b7b422a6c"`

	RepoId       string `json:"repoId" gorm:"not null"`       // RepoId 仓库 id 或者 path(local vcs)
	RepoFullName string `json:"repoFullName" gorm:"not null"` // 完整的仓库名称
	RepoRevision string `json:"repoRevision" gorm:"size:64;default:'master'" example:"master"`

	// 云模板的 repoAddr 和 repoToken 字段可以为空，若为空则在创建 task 时会查询 vcs 获取
	// 提供这两个字段主要是为了后续支持直接添加 git 地址和 token 来创建云模板
	RepoAddr  string `json:"repoAddr" gorm:"not null" example:"https://github.com/user/project.git"` // RepoAddr 仓库地址(完整 url 或者项目 path)
	RepoToken string `json:"repoToken" gorm:"size:128" `                                             // RepoToken 若为空则使用 vcs 的 token

	Status     string `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:状态"`
	CreatorId  Id     `json:"creatorId" gorm:"size:32;not null;comment:创建人"`
	Workdir    string `json:"workdir" gorm:"default:''" example:"aws"` // 基于项目根目录的相对路径, 默认为空
	TfVarsFile string `json:"tfVarsFile" gorm:"default:''"`            // Terraform 变量文件路径

	// 要执行的 ansible playbook 文件(基于 workdir 的相对路径)
	Playbook     string `json:"playbook" gorm:"default:''" example:"ansbile/playbook.yml"`
	PlayVarsFile string `json:"playVarsFile" gorm:"default:''"` // Ansible 变量文件路径

	LastScanTaskId Id `json:"lastScanTaskId" gorm:"size:32"` // 最后一次策略扫描任务 id

	TfVersion string `json:"tfVersion" gorm:"default:''"` // 模版使用的terraform版本号

	// 触发器设置
	Triggers pq.StringArray `json:"triggers" gorm:"type:text" swaggertype:"array,string"` // 触发器。commit（每次推送执行合规检测）
}

func (Template) TableName() string {
	return "iac_template"
}

func (Template) NewId() Id {
	return NewId("tpl")
}

func (t Template) Migrate(sess *db.Session) (err error) {
	if err = sess.RemoveIndex("iac_template", "unique__project__tpl__name"); err != nil {
		return err
	}
	if err = t.AddUniqueIndex(sess, "unique__org__tpl__name", "org_id", "name"); err != nil {
		return err
	}
	return nil
}
