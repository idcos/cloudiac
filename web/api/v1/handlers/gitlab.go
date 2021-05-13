package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type GitLab struct {}


func (GitLab) ListRepos(c *ctx.GinRequestCtx) {
	form := forms.GetGitProjectsForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	if form.Type == "gitlab"{
		c.JSONResult(apps.ListOrganizationRepos(&form))
	} else if form.Type == "gitea" {
		c.JSONResult(apps.ListGiteaOrganizationRepos(&form))
	}

}


func (GitLab) ListBranches(c *ctx.GinRequestCtx) {
	form := forms.GetGitBranchesForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	if form.Type == "gitlab" {
		c.JSONResult(apps.ListRepositoryBranches(&form))
	} else if form.Type == "gitea" {
		c.JSONResult(apps.ListGiteaRepoBranches(&form))
	}

}

func (GitLab) GetReadmeContent(c *ctx.GinRequestCtx) {
	form := forms.GetReadmeForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	if form.Type == "gitlab" {
		c.JSONResult(apps.GetReadmeContent(&form))
	} else if form.Type == "gitea" {
		c.JSONResult(apps.GetGiteaReadme(&form))
	}

}
