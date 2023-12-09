// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package resps

import (
	"cloudiac/portal/models"
	"cloudiac/portal/models/desensitize"
)

type SearchTemplateResp struct {
	CreatedAt           models.Time `json:"createdAt"` // 创建时间
	UpdatedAt           models.Time `json:"updatedAt"` // 更新时间
	Id                  models.Id   `json:"id"`
	Name                string      `json:"name"`
	Description         models.Text `json:"description"`
	ActiveEnvironment   *int        `json:"activeEnvironment,omitempty"` // 定义为 *int，以区分数值 0 和未赋值
	RelationEnvironment *int        `json:"relationEnvironment,omitempty"`
	RepoRevision        string      `json:"repoRevision"`
	Creator             string      `json:"creator"`
	RepoId              string      `json:"repoId"`
	VcsId               string      `json:"vcsId"`
	RepoAddr            string      `json:"repoAddr"`
	TplType             string      `json:"tplType" `
	RepoFullName        string      `json:"repoFullName"`
	NewRepoAddr         string      `json:"newRepoAddr"`
	VcsAddr             string      `json:"vcsAddr"`
	PolicyEnable        bool        `json:"policyEnable"`
	PolicyStatus        string      `json:"policyStatus"`
	IsDemo              bool        `json:"isDemo"`
}

type TemplateDetailResp struct {
	*models.Template
	Variables   []desensitize.Variable `json:"variables"`
	ProjectList []models.Id            `json:"projectId"`
	PolicyGroup []string               `json:"policyGroup"`
}
