package handlers

import (
	"cloudiac/apps"
	"cloudiac/consts/e"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
	"cloudiac/services"
)

type Vcs struct {
	ctrl.BaseController
}

func (Vcs) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateVcs(c.ServiceCtx(), form))
}

func (Vcs) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchVcs(c.ServiceCtx(), form))
}

func (Vcs) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateVcs(c.ServiceCtx(), form))
}

func (Vcs) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteVcs(c.ServiceCtx(), form))
}

func ListEnableVcs(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.ListEnableVcs(c.ServiceCtx()))
}

func (Vcs) ListRepos(c *ctx.GinRequestCtx) {
	form := forms.GetGitProjectsForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	vcs, err := services.QueryVcsByVcsId(form.VcsId, c.ServiceCtx().Tx())
	if err != nil {
		c.JSONResult(nil,e.New(e.DBError, err))
		return
	}
	if vcs.VcsType == "gitlab"{
		c.JSONResult(apps.ListOrganizationRepos(vcs, &form))
	} else if vcs.VcsType == "gitea" {
		c.JSONResult(apps.ListGiteaOrganizationRepos(vcs, &form))
	}

}


func (Vcs) ListBranches(c *ctx.GinRequestCtx) {
	form := forms.GetGitBranchesForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	vcs, err := services.QueryVcsByVcsId(form.VcsId, c.ServiceCtx().Tx())
	if err != nil {
		c.JSONResult(nil,e.New(e.DBError, err))
		return
	}
	if vcs.VcsType == "gitlab" {
		c.JSONResult(apps.ListRepositoryBranches(vcs, &form))
	} else if vcs.VcsType == "gitea" {
		c.JSONResult(apps.ListGiteaRepoBranches(vcs, &form))
	}

}

func (Vcs) GetReadmeContent(c *ctx.GinRequestCtx) {
	form := forms.GetReadmeForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	vcs, err := services.QueryVcsByVcsId(form.VcsId, c.ServiceCtx().Tx())
	if err != nil {
		c.JSONResult(nil,e.New(e.DBError, err))
		return
	}
	if vcs.VcsType == "gitlab" {
		c.JSONResult(apps.GetReadmeContent(vcs, &form))
	} else if vcs.VcsType == "gitea" {
		c.JSONResult(apps.GetGiteaReadme(vcs, &form))
	}

}

func TemplateTfvarsSearch(c *ctx.GinRequestCtx){
	form := forms.TemplateTfvarsSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	vcs, err := services.QueryVcsByVcsId(form.VcsId, c.ServiceCtx().Tx())
	if err != nil {
		c.JSONResult(nil,e.New(e.DBError, err))
		return
	}
	c.JSONResult(apps.TemplateTfvarsSearch(vcs, &form))
}

func TemplatePlaybookSearch(c *ctx.GinRequestCtx){
	form := forms.TemplatePlaybookSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	vcs, err := services.QueryVcsByVcsId(form.VcsId, c.ServiceCtx().Tx())
	if err != nil {
		c.JSONResult(nil,e.New(e.DBError, err))
		return
	}
	c.JSONResult(apps.TemplatePlaybookSearch(vcs, &form))
}