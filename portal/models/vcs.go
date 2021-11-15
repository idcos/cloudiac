// Copyright 2021 CloudJ Company Limited. All rights reserved.

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
)

type Vcs struct {
	BaseModel
	OrgId     Id     `json:"orgId" gorm:"size:32;not null"` // 默认仓库的 orgId 为 ""
	ProjectId Id     `json:"projectId" gorm:"size:32;not null"`
	Name      string `json:"name" gorm:"not null;comment:vcs名称"`
	Status    string `json:"status" gorm:"type:enum('enable','disable');default:'enable';comment:vcs状态"`
	VcsType   string `json:"vcsType" gorm:"not null;comment:vcs代码库类型"`
	Address   string `json:"address" gorm:"not null;comment:vcs代码库地址"`
	VcsToken  string `json:"vcsToken" gorm:"not null; comment:代码库的token值"`
}

func (Vcs) TableName() string {
	return "iac_vcs"
}

func (Vcs) NewId() Id {
	return NewId("vcs")
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
