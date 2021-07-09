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
	PageForm

	Name                   string    `form:"name" json:"name" binding:"required,gte=2,lte=32"`
	TplType     		   string    `form:"tplType" json:"tplType" binding:"required"`
	OrgId				   string    `form:"orgId" json:"orgId" binding:"required"`
	Description            string    `form:"description" json:"Description" binding:""`
	RepoId                 string    `form:"repoId" json:"repoId" binding:""`
	RepoAddr               string    `form:"repoAddr" json:"repoAddr" binding:""`
	RepoRevision           string    `form:"repoRevision" json:"repoRevision" binding:""`
	Extra                  string    `form:"extra" json:"extra"`
	Workdir				   string    `form:"workdir" json:"workdor"`
	VcsId                  models.Id `form:"vcsId" json:"vcsId" binding:"required"`
	Playbook               string    `json:"playbook" form:"playbook"`
	Status				   string    `json:"status" form:"status"`
	CreatorId			   string    `json:"creatorId" form:"creatorId" binding:"required"`
	RunnerId			   string    `json:"runnerId" form:"runnerId" binding:"required"`
}


type SearchTemplateForm struct {
	PageForm

	Q          string `form:"q" json:"q" binding:""`
	Status     string `form:"status" json:"status"`
}

type UpdateTemplateForm struct {
	PageForm
	Id                     models.Id `form:"id" json:"id" binding:"required"`
	Name                   string    `form:"name" json:"name"`
	Description            string    `form:"description" json:"Description"`
	Extra                  string    `form:"extra" json:"extra"`
	Status                 string    `form:"status" json:"status"`
	Workdir				   string    `form:"workdir" json:"workdor"`
	RunnerId			   string    `json:"runnerId" form:"runnerId"`
	Playbook               string    `json:"playbook" form:"playbook"`
}

type DeleteTemplateForm struct {
	BaseForm
	Id     models.Id `form:"id" json:"id" binding:"required`
}


type DetailTemplateForm struct {
	PageForm
	Id models.Id `form:"id" json:"id" binding:"required"`
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
	PageForm
	RepoId     string    `json:"repoId" form:"repoId"`
	RepoBranch string    `json:"repoBranch" form:"repoBranch" `
	RepoType   string    `json:"repoType" form:"repoType" `
	VcsId      models.Id `json:"vcsId" form:"vcsId"`
}

type TemplateVariableSearchForm struct {
	PageForm
	RepoId     string    `json:"repoId" form:"repoId" `
	RepoBranch string    `json:"repoBranch" form:"repoBranch" `
	RepoType   string    `json:"repoType" form:"repoType" `
	VcsId      models.Id `json:"vcsId" form:"vcsId"`
}

type TemplatePlaybookSearchForm struct {
	PageForm
	RepoId     string    `json:"repoId" form:"repoId" `
	RepoBranch string    `json:"repoBranch" form:"repoBranch" `
	RepoType   string    `json:"repoType" form:"repoType" `
	VcsId      models.Id `json:"vcsId" form:"vcsId"`
}
