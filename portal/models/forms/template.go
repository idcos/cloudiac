// Copyright 2021 CloudJ Company Limited. All rights reserved.

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

	Name              string      `form:"name" json:"name" binding:"required,gte=2,lte=64"`
	Description       string      `form:"description" json:"description" binding:""`
	RepoId            string      `form:"repoId" json:"repoId" binding:"required"`
	RepoRevision      string      `form:"repoRevision" json:"repoRevision" binding:""`
	Extra             string      `form:"extra" json:"extra"`
	Workdir           string      `form:"workdir" json:"workdir"`
	VcsId             models.Id   `form:"vcsId" json:"vcsId" binding:"required"`
	Playbook          string      `json:"playbook" form:"playbook"`
	PlayVarsFile      string      `json:"playVarsFile" form:"playVarsFile"`
	TfVarsFile        string      `form:"tfVarsFile" json:"tfVarsFile"`
	Variables         []Variables `json:"variables" form:"variables" `
	DeleteVariablesId []string    `json:"deleteVariablesId" form:"deleteVariablesId" ` //变量id
	ProjectId         []models.Id `form:"projectId" json:"projectId"`                  // 项目ID
	TfVersion         string      `form:"tfVersion" json:"tfVersion"`                  // 模版使用terraform版本号
	VarGroupIds       []models.Id `json:"varGroupIds" form:"varGroupIds" `
	DelVarGroupIds    []models.Id `json:"delVarGroupIds" form:"delVarGroupIds" `
}

type SearchTemplateForm struct {
	PageForm

	Q      string `form:"q" json:"q" binding:""`
	Status string `form:"status" json:"status"`
}

type UpdateTemplateForm struct {
	BaseForm
	Id                models.Id   `uri:"id" form:"id" json:"id" binding:"required"`
	Name              string      `form:"name" json:"name"`
	Description       string      `form:"description" json:"description"`
	Status            string      `form:"status" json:"status"`
	Workdir           string      `form:"workdir" json:"workdir"`
	RunnerId          string      `json:"runnerId" form:"runnerId"`
	Playbook          string      `json:"playbook" form:"playbook"`
	PlayVarsFile      string      `json:"playVarsFile" form:"playVarsFile"`
	TfVarsFile        string      `form:"tfVarsFile" json:"tfVarsFile"`
	Variables         []Variables `json:"variables" form:"variables" `
	DeleteVariablesId []string    `json:"deleteVariablesId" form:"deleteVariablesId" ` //变量id
	ProjectId         []models.Id `form:"projectId" json:"projectId"`
	RepoRevision      string      `form:"repoRevision" json:"repoRevision" binding:""`
	VcsId             models.Id   `form:"vcsId" json:"vcsId" binding:""`
	RepoId            string      `form:"repoId" json:"repoId" binding:""`
	TfVersion         string      `form:"tfVersion" json:"tfVersion" binding:""`
	VarGroupIds       []models.Id `json:"varGroupIds" form:"varGroupIds" `
	DelVarGroupIds    []models.Id `json:"delVarGroupIds" form:"delVarGroupIds" `
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

type TemplateTfvarsSearchForm struct {
	BaseForm
	RepoId       string    `json:"repoId" form:"repoId" binding:"required"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision" binding:"required"`
	RepoType     string    `json:"repoType" form:"repoType" `
	VcsId        models.Id `uri:"id"`
	TplChecks    bool      `json:"tplChecks" form:"tplChecks"`
	Path         string    `json:"path" form:"path"`
}

type TemplateVariableSearchForm struct {
	BaseForm
	RepoId       string    `json:"repoId" form:"repoId" binding:"required"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision" binding:"required"`
	RepoType     string    `json:"repoType" form:"repoType" `
	VcsId        models.Id `json:"vcsId" form:"vcsId" binding:"required"`
}

type TemplatePlaybookSearchForm struct {
	BaseForm
	RepoId       string    `json:"repoId" form:"repoId" binding:"required"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision" binding:"required"`
	RepoType     string    `json:"repoType" form:"repoType" `
	VcsId        models.Id `uri:"id"`
}

type TemplateTfVersionSearchForm struct {
	BaseForm
	VcsId     models.Id `json:"vcsId" form:"vcsId"`
	VcsBranch string    `json:"vcsBranch" form:"vcsBranch"`
	RepoId    string    `json:"repoId" form:"repoId"`
}

type TemplateChecksForm struct {
	BaseForm
	Name         string    `json:"name" form:"name" form:"name"`
	RepoId       string    `json:"repoId" form:"repoId"`
	RepoRevision string    `json:"repoRevision" form:"repoRevision"`
	RepoType     string    `json:"repoType" form:"repoType" `
	VcsId        models.Id `json:"vcsId" form:"vcsId"`
	Path         string    `json:"path" form:"path"`
}
