package models

import (
	"cloudiac/portal/libs/db"
)

type PolicyGroup struct {
	SoftDeleteModel

	OrgId     Id `json:"orgId" gorm:"size:32;comment:组织ID" example:"org-c3lcrjxczjdywmk0go90"`
	CreatorId Id `json:"creatorId" gorm:"size:32;not null;comment:创建人ID" example:"u-c3lcrjxczjdywmk0go90"`

	Name        string `json:"name" gorm:"not null;size:128;comment:策略组名称" example:"安全合规策略组"`
	Description string `json:"description" gorm:"type:text;comment:描述" example:"本组包含对于安全合规的检查策略"`
	Enabled     bool   `json:"enabled" gorm:"default:true;comment:是否启用" example:"true"`
	Source      string `json:"source" gorm:"type:enum('vcs','registry');comment:来源：VCS/Registry"`
	VcsId       Id     `json:"vcsId" gorm:"size:32;not null;comment:VCS ID"`
	RepoId      string `json:"repoId" gorm:"size:128;not null;comment:VCS 仓库ID"`
	GitTags     string `json:"gitTags" gorm:"size:128;comment:Git 版本标签：\"v1.0.0\""`
	Branch      string `json:"branch" gorm:"size:128;comment:分支"`
	CommitId    string `json:"commitId" gorm:"size:128;not null;当前 git commit id"`
	UseLatest   bool   `json:"useLatest" gorm:"default:false;comment:是否跟踪最新版本，如果从分支导入，默认为true" example:"true"`
	Version     string `json:"version" gorm:"size:32;not null;策略组版本：\"1.0.0\""`
	Dir         string `json:"dir" gorm:"default:\"/\";comment:策略组目录，默认为根目录：/"`
	Label       string `json:"label" gorm:"size:128;comment:策略组标签，多个值以 , 分隔"`
}

func (PolicyGroup) TableName() string {
	return "iac_policy_group"
}

func (g *PolicyGroup) CustomBeforeCreate(*db.Session) error {
	if g.Id == "" {
		g.Id = NewId("pog")
	}
	return nil
}

func (g PolicyGroup) Migrate(sess *db.Session) error {
	if err := g.AddUniqueIndex(sess, "unique__name", "name"); err != nil {
		return err
	}
	return nil
}
