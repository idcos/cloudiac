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

	Name         string    `form:"name" json:"name" binding:"required,gte=2,lte=32"`
	TplType      string    `form:"tplType" json:"tplType" binding:"required"`
	Description  string    `form:"description" json:"description" binding:""`
	RepoId       string    `form:"repoId" json:"repoId" binding:""`
	RepoAddr     string    `form:"repoAddr" json:"repoAddr" binding:""`
	RepoRevision string    `form:"repoRevision" json:"repoRevision" binding:""`
	Extra        string    `form:"extra" json:"extra"`
	Workdir      string    `form:"workdir" json:"workdir"`
	VcsId        models.Id `form:"vcsId" json:"vcsId" binding:"required"`
	Playbook     string    `json:"playbook" form:"playbook"`
	PlayVarsFile string    `json:"playVarsFile" form:"playVarsFile"`
	TfVarsFile   string    `form:"tfVarsFile" json:"tfVarsFile"`
	Variables         []Variables `json:"variables" form:"variables" `
	DeleteVariablesId []string    `json:"deleteVariablesId" form:"deleteVariablesId" ` //变量id
	ProjectId    []models.Id `form:"projectId" json:"projectId"` // 项目ID
}

type SearchTemplateForm struct {
	PageForm

	Q      string `form:"q" json:"q" binding:""`
	Status string `form:"status" json:"status"`
}

type UpdateTemplateForm struct {
	BaseForm
	Id           models.Id `form:"id" json:"id" binding:"required"`
	Name         string    `form:"name" json:"name"`
	Description  string    `form:"description" json:"description"`
	Extra        string    `form:"extra" json:"extra"`
	Status       string    `form:"status" json:"status"`
	Workdir      string    `form:"workdir" json:"workdor"`
	RunnerId     string    `json:"runnerId" form:"runnerId"`
	Playbook     string    `json:"playbook" form:"playbook"`
	PlayVarsFile string    `json:"playVarsFile" form:"playVarsFile"`
	TfVarsFile   string    `form:"tfVarsFile" json:"tfVarsFile"`
	Variables         []Variables `json:"variables" form:"variables" `
	DeleteVariablesId []string    `json:"deleteVariablesId" form:"deleteVariablesId" ` //变量id
}

type DeleteTemplateForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" binding:"required" swaggerignore:"true"`
}

type DetailTemplateForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" binding:"required" swaggerignore:"true"`
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
	RepoId     string    `json:"repoId" form:"repoId" binding:"required"`
	RepoBranch string    `json:"repoBranch" form:"repoBranch" binding:"required"`
	RepoType   string    `json:"repoType" form:"repoType" `
	VcsId      models.Id `json:"vcsId" form:"vcsId" binding:"required"`
}

type TemplateVariableSearchForm struct {
	BaseForm
	RepoId     string    `json:"repoId" form:"repoId" binding:"required"`
	RepoBranch string    `json:"repoBranch" form:"repoBranch" binding:"required"`
	RepoType   string    `json:"repoType" form:"repoType" `
	VcsId      models.Id `json:"vcsId" form:"vcsId" binding:"required"`
}

type TemplatePlaybookSearchForm struct {
	BaseForm
	RepoId     string    `json:"repoId" form:"repoId" binding:"required"`
	RepoBranch string    `json:"repoBranch" form:"repoBranch" binding:"required"`
	RepoType   string    `json:"repoType" form:"repoType" `
	VcsId      models.Id `json:"vcsId" form:"vcsId" binding:"required"`
}
