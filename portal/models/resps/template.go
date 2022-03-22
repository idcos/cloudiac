package resps

import "cloudiac/portal/models"

type SearchTemplateResp struct {
	CreatedAt           models.Time `json:"createdAt"` // 创建时间
	UpdatedAt           models.Time `json:"updatedAt"` // 更新时间
	Id                  models.Id   `json:"id"`
	Name                string      `json:"name"`
	Description         string      `json:"description"`
	ActiveEnvironment   int         `json:"activeEnvironment"`
	RelationEnvironment int         `json:"relationEnvironment"`
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
}

type TemplateDetailResp struct {
	*models.Template
	Variables   []models.Variable `json:"variables"`
	ProjectList []models.Id       `json:"projectId"`
	PolicyGroup []string          `json:"policyGroup"`
}
