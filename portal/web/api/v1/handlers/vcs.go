package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
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
	c.JSONResult(apps.ListRepos(c.ServiceCtx(), &form))
}

func (Vcs) ListBranches(c *ctx.GinRequestCtx) {
	form := forms.GetGitBranchesForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ListRepoBranches(c.ServiceCtx(), &form))
}

func (Vcs) GetReadmeContent(c *ctx.GinRequestCtx) {
	form := forms.GetReadmeForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.GetReadme(c.ServiceCtx(), &form))
}

func TemplateTfvarsSearch(c *ctx.GinRequestCtx) {
	form := forms.TemplateTfvarsSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsTfVarsSearch(c.ServiceCtx(), &form))
}


func TemplateVariableSearch(c *ctx.GinRequestCtx) {
	form := forms.TemplateVariableSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsVariableSearch(c.ServiceCtx(), &form))
}


func TemplatePlaybookSearch(c *ctx.GinRequestCtx) {
	form := forms.TemplatePlaybookSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsPlaybookSearch(c.ServiceCtx(), &form))
}
