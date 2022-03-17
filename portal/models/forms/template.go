// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type Var struct {
	Id          string `form:"id" json:"id" binding:"required"`
	Key         string `form:"key" json:"key" binding:"required"`
	Value       string `form:"value" json:"value" binding:"required"`
	IsSecret    *bool  `form:"isSecret" json:"isSecret" binding:"required,default:false"`
	Type        string `form:"type" json:"type" binding:"required,default:env"`
	Description string `form:"description" json:"description" binding:""`
}

type CreateTemplateForm struct {
	BaseForm

	Name         string      `form:"name" json:"name" binding:"required,gte=2,lte=64"`
	Description  string      `form:"description" json:"description" binding:""`
	RepoId       string      `form:"repoId" json:"repoId" binding:"required"`
	RepoFullName string      `form:"repoFullName" json:"repoFullName" binding:"required"`
	RepoRevision string      `form:"repoRevision" json:"repoRevision" binding:""`
	Extra        string      `form:"extra" json:"extra"`
	Workdir      string      `form:"workdir" json:"workdir"`
	VcsId        models.Id   `form:"vcsId" json:"vcsId" binding:"required"`
	Playbook     string      `json:"playbook" form:"playbook"`
	PlayVarsFile string      `json:"playVarsFile" form:"playVarsFile"`
	TfVarsFile   string      `form:"tfVarsFile" json:"tfVarsFile"`
	ProjectId    []models.Id `form:"projectId" json:"projectId"` // 项目ID
	TfVersion    string      `form:"tfVersion" json:"tfVersion"` // 模版使用terraform版本号

	Variables []Variable `json:"variables" form:"variables" `

	VarGroupIds    []models.Id `json:"varGroupIds" form:"varGroupIds" `
	DelVarGroupIds []models.Id `json:"delVarGroupIds" form:"delVarGroupIds" `
	PolicyEnable   bool        `json:"policyEnable" form:"policyEnable"` // 是否开启合规检测
	PolicyGroup    []models.Id `json:"policyGroup" form:"policyGroup"`   // 绑定的合规策略组
	TplTriggers    []string    `json:"tplTriggers" form:"tplTriggers"`   // 分之推送自动触发合规 例如 ["commit"]

	KeyId models.Id `form:"keyId" json:"keyId" binding:""` // 部署密钥ID

}

type SearchTemplateForm struct {
	PageForm

	Q      string `form:"q" json:"q" binding:""`
	Status string `form:"status" json:"status"`
}

type UpdateTemplateForm struct {
	BaseForm
	Id           models.Id   `uri:"id" form:"id" json:"id" binding:"required"`
	Name         string      `form:"name" json:"name"`
	Description  string      `form:"description" json:"description"`
	Status       string      `form:"status" json:"status"`
	Workdir      string      `form:"workdir" json:"workdir"`
	RunnerId     string      `json:"runnerId" form:"runnerId"`
	Playbook     string      `json:"playbook" form:"playbook"`
	PlayVarsFile string      `json:"playVarsFile" form:"playVarsFile"`
	TfVarsFile   string      `form:"tfVarsFile" json:"tfVarsFile"`
	ProjectId    []models.Id `form:"projectId" json:"projectId"`
	RepoRevision string      `form:"repoRevision" json:"repoRevision" binding:""`
	VcsId        models.Id   `form:"vcsId" json:"vcsId" binding:""`
	RepoId       string      `form:"repoId" json:"repoId" binding:""`
	RepoFullName string      `form:"repoFullName" json:"repoFullName" binding:""`
	TfVersion    string      `form:"tfVersion" json:"tfVersion" binding:""`

	Variables []Variable `json:"variables" form:"variables" `

	VarGroupIds    []models.Id `json:"varGroupIds" form:"varGroupIds" `
	DelVarGroupIds []models.Id `json:"delVarGroupIds" form:"delVarGroupIds" `
	PolicyEnable   bool        `json:"policyEnable" form:"policyEnable"` // 是否开启合规检测
	PolicyGroup    []models.Id `json:"policyGroup" form:"policyGroup"`   // 绑定的合规策略组
	TplTriggers    []string    `json:"tplTriggers" form:"tplTriggers"`   // 分之推送自动触发合规 例如 ["commit"]
	KeyId          models.Id   `form:"keyId" json:"keyId" binding:""`    // 部署密钥ID
}

type DeleteTemplateForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" binding:"required" swaggerignore:"true"`
}

type DetailTemplateForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" binding:"required" swaggerignore:"true"`
	// TODO 返回要返回 projectId
}

type OpenApiDetailTemplateForm struct {
	PageForm
	Id models.Id `form:"id" json:"id" binding:"required"`
}

type OverviewTemplateForm struct {
	PageForm
	Id models.Id `form:"id" json:"id" binding:"required"`
}

type RepoFileSearchForm struct {
	BaseForm
	RepoId       string    `json:"repoId" form:"repoId" binding:"required"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision" binding:"required"`
	VcsId        models.Id `uri:"id"`
	Workdir      string    `json:"workdir" form:"workdir" `
}

type TemplateVariableSearchForm struct {
	BaseForm
	RepoId       string    `json:"repoId" form:"repoId" binding:"required"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision" binding:"required"`
	VcsId        models.Id `uri:"id"`
	Workdir      string    `json:"workdir" form:"workdir" `
}

type TemplateTfVersionSearchForm struct {
	BaseForm
	VcsId     models.Id `json:"vcsId" form:"vcsId"`
	VcsBranch string    `json:"vcsBranch" form:"vcsBranch"`
	RepoId    string    `json:"repoId" form:"repoId"`
}

type TemplateChecksForm struct {
	BaseForm
	Name         string    `json:"name" form:"name"`
	RepoId       string    `json:"repoId" form:"repoId"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision"`
	VcsId        models.Id `json:"vcsId" form:"vcsId"`
	Workdir      string    `json:"workdir" form:"workdir"`
	TemplateId   models.Id `json:"templateId" form:"templateId"`
	TfVarsFile   string    `json:"tfVarsFile" form:"tfVarsFile"`
	Playbook     string    `json:"playbook" form:"playbook"`
}
