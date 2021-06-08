package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
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
	c.JSONResult(apps.GetReadme(c.ServiceCtx(),&form))
}

func TemplateTfvarsSearch(c *ctx.GinRequestCtx){
	form := forms.TemplateTfvarsSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsTfVarsSearch(c.ServiceCtx(), &form))
}

// TemplateVariableSearch 查询云模板TF参数
// @Tags 云模板
// @Description 云模板参数接口
// @Accept application/json
// @Param repoId formData int true "仓库id"
// @Param repoBranch formData int true "分支"
// @Param vcsId formData int true "vcsID"
// @router /api/v1/template/variable/search [get]
func TemplateVariableSearch(c *ctx.GinRequestCtx){
	form := forms.TemplateVariableSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsVariableSearch(c.ServiceCtx(), &form))
}

//TemplatePlaybookSearch
// @Tags playbook列表查询
// @Description  playbook列表接口
// @Accept application/json
// @Param repoId formData int true "仓库id"
// @Param repoBranch formData int true "分支"
// @Param vcsId formData int true "vcsID"
// @router /api/v1/template/playbook/search [get]
func TemplatePlaybookSearch(c *ctx.GinRequestCtx){
	form := forms.TemplatePlaybookSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsPlaybookSearch(c.ServiceCtx(), &form))
}


