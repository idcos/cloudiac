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
	Description            string    `form:"description" json:"Description" binding:""`
	RepoId                 string    `form:"repoId" json:"repoId" binding:""`
	RepoAddr               string    `form:"repoAddr" json:"repoAddr" bingding:""`
	RepoBranch             string    `form:"repoBranch" json:"repoBranch" bingding:""`
	SaveState              *bool     `form:"saveState" json:"saveState"`
	Vars                   []Var     `form:"vars" json:"vars"`
	Varfile                string    `form:"varfile" json:"varfile"`
	Extra                  string    `form:"extra" json:"extra"`
	Timeout                int64     `form:"timeout" json:"timeout"`
	VcsId                  models.Id `json:"vcsId"`
	DefaultRunnerAddr      string    `json:"defaultRunnerAddr" `
	DefaultRunnerPort      uint      `json:"defaultRunnerPort" `
	DefaultRunnerServiceId string    `json:"defaultRunnerServiceId"`
	Playbook               string    `json:"playbook" form:"playbook" `
}

type SearchTemplateForm struct {
	PageForm

	Q          string `form:"q" json:"q" binding:""`
	Status     string `form:"status" json:"status"`
	TaskStatus string `json:"taskStatus" form:"taskStatus" `
}

type UpdateTemplateForm struct {
	PageForm
	Id                     models.Id `form:"id" json:"id" binding:"required"`
	Name                   string    `form:"name" json:"name"`
	Description            string    `form:"description" json:"Description"`
	SaveState              bool      `form:"saveState" json:"saveState"`
	Vars                   []Var     `form:"vars" json:"vars"`
	Varfile                string    `form:"varfile" json:"varfile"`
	Extra                  string    `form:"extra" json:"extra"`
	Timeout                int       `form:"timeout" json:"timeout"`
	Status                 string    `form:"status" json:"status"`
	DefaultRunnerAddr      string    `json:"defaultRunnerAddr" grom:"not null;comment:'默认runner地址'"`
	DefaultRunnerPort      uint      `json:"defaultRunnerPort" grom:"not null;comment:'默认runner端口'"`
	DefaultRunnerServiceId string    `json:"defaultRunnerServiceId" grom:"not null;comment:'默认runner-consul-serviceId'"`
	Playbook               string    `json:"playbook" form:"playbook" `
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
