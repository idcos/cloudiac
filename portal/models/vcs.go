// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/common"
	"cloudiac/portal/libs/db"
	"cloudiac/utils"
)

const (
	VcsGitlab = common.VcsGitlab
	VcsGitea  = common.VcsGitea
	VcsGitee  = common.VcsGitee
	VcsGithub = common.VcsGithub
	// git clone 鉴权时使用的user 默认为token
	RepoUser = "token"
)

type Vcs struct {
	BaseModel
	OrgId     Id     `json:"orgId" gorm:"size:32;not null"` // 默认仓库的 orgId 为 ""
	ProjectId Id     `json:"projectId" gorm:"size:32;not null"`
	Name      string `json:"name" gorm:"not null;comment:vcs名称"`
	Status    string `json:"status" gorm:"default:'enable';comment:vcs状态"` // type:enum('enable','disable');
	VcsType   string `json:"vcsType" gorm:"not null;comment:vcs代码库类型"`
	Address   string `json:"address" gorm:"not null;comment:vcs代码库地址"`
	VcsToken  string `json:"vcsToken" gorm:"not null; comment:代码库的token值"`

	IsDemo bool `json:"isDemo"`
}

func (Vcs) TableName() string {
	return "iac_vcs"
}

func (Vcs) NewId() Id {
	return NewId("vcs")
}

//go:generate go run cloudiac/code-gen/desenitize Vcs ./desensitize/
func (v *Vcs) Desensitize() Vcs {
	rv := Vcs{}
	utils.DeepCopy(&rv, v)
	rv.VcsToken = ""
	return rv
}

func (v Vcs) Migrate(sess *db.Session) (err error) {
	if err = v.AddUniqueIndex(sess, "unique__org_vcs_name", "org_id", "name"); err != nil {
		return err
	}
	return nil
}

func (v *Vcs) DecryptToken() (string, error) {
	return utils.DecryptSecretVar(v.VcsToken)
}

type VcsPr struct {
	AutoUintIdModel

	PrId   int `json:"prId" form:"prId" `
	TaskId Id  `json:"taskId" form:"taskId" `
	EnvId  Id  `json:"envId" form:"envId" `
	VcsId  Id  `json:"vcsId" form:"vcsId" `
}

func (VcsPr) TableName() string {
	return "iac_vcs_pr"
}
