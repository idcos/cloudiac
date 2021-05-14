package handlers

import (
	"cloudiac/apps"
	"cloudiac/configs"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type GitLab struct {}


func (GitLab) ListRepos(c *ctx.GinRequestCtx) {
	form := forms.GetGitProjectsForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	conf := configs.Get()
	if conf.Gitlab.Type == "gitlab"{
		c.JSONResult(apps.ListOrganizationRepos(c.ServiceCtx(), &form))
	} else if conf.Gitlab.Type == "gitea" {
		c.JSONResult(apps.ListGiteaOrganizationRepos(&form))
	}

}

func (GitLab) ListBranches(c *ctx.GinRequestCtx) {
	form := forms.GetGitBranchesForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	conf := configs.Get()
	if conf.Gitlab.Type == "gitlab" {
		c.JSONResult(apps.ListRepositoryBranches(c.ServiceCtx(), &form))
	} else if conf.Gitlab.Type == "gitea" {
		c.JSONResult(apps.ListGiteaRepoBranches(&form))
	}

}

func (GitLab) GetReadmeContent(c *ctx.GinRequestCtx) {
	form := forms.GetReadmeForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	conf := configs.Get()
	if conf.Gitlab.Type == "gitlab" {
		c.JSONResult(apps.GetReadmeContent(c.ServiceCtx(), &form))
	} else if conf.Gitlab.Type == "gitea" {
		c.JSONResult(apps.GetGiteaReadme(&form))
	}

}

func TemplateTfvarsSearch(c *ctx.GinRequestCtx){
	form := forms.TemplateTfvarsSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.TemplateTfvarsSearch(c.ServiceCtx(), &form))
}
