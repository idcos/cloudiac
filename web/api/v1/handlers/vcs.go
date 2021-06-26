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

// Create 创建VCS来源
// @Summary 创建VCS来源
// @Description 创建VCS来源
// @Tags VCS
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.CreateVcsForm true "VCS信息"
// @Success 200 {object} models.Vcs
// @Router /vcs/create [post]
func (Vcs) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateVcs(c.ServiceCtx(), form))
}

// Search 查询VCS列表
// @Summary 查询VCS列表
// @Description 查询VCS列表
// @Tags VCS
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param q query string false "模糊搜索"
// @Param status query string false "VCS状态"
// @Success 200 {object} models.Vcs
// @Router /vcs/search [get]
func (Vcs) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchVcs(c.ServiceCtx(), form))
}

// Update 修改VCS信息
// @Summary 修改VCS信息
// @Description 修改VCS信息
// @Tags VCS
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.UpdateVcsForm true "vcs信息"
// @Success 200 {object} models.Vcs
// @Router /vcs/update [put]
func (Vcs) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateVcs(c.ServiceCtx(), form))
}

// Delete 删除VCS
// @Summary 删除VCS
// @Description 删除VCS
// @Tags VCS
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.DeleteVcsForm true "vcs信息"
// @Success 200
// @Router /vcs/delete [delete]
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

// ListRepos 查询repo列表
// @Summary 查询repo列表
// @Description 查询repo列表
// @Tags VCS
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param q query string false "模糊搜索"
// @Param vcsId query int true "vcsId"
// @Success 200 {object} vcsrv.Projects
// @Router /vcs/repo/search [get]
func (Vcs) ListRepos(c *ctx.GinRequestCtx) {
	form := forms.GetGitProjectsForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ListRepos(c.ServiceCtx(), &form))
}

// ListBranches 查询repo下分支列表
// @Summary 查询repo下分支列表
// @Description 查询repo下分支列表
// @Tags VCS
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param repoId query string true "repoId"
// @Param vcsId query int true "vcsId"
// @Success 200 {object} []apps.Branches
// @Router /vcs/branch/search [get]
func (Vcs) ListBranches(c *ctx.GinRequestCtx) {
	form := forms.GetGitBranchesForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ListRepoBranches(c.ServiceCtx(), &form))
}

// GetReadmeContent 查询repo下Readme
// @Summary 查询repo下Readme
// @Description 查询repo下Readme
// @Tags VCS
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param repoId query string true "repoId"
// @Param branch query string true "repo branch"
// @Param vcsId query int true "vcsId"
// @Success 200 {object}  models.FileContent
// @Router /vcs/readme [get]
func (Vcs) GetReadmeContent(c *ctx.GinRequestCtx) {
	form := forms.GetReadmeForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.GetReadme(c.ServiceCtx(),&form))
}

// TemplateTfvarsSearch 查询repo下tfvar文件列表
// @Summary 查询repo下tfvar文件列表
// @Description 查询repo下tfvar文件列表
// @Tags 云模板
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param repoId query string true "repoId"
// @Param repoBranch query string true "repo branch"
// @Param vcsId query int true "vcsId"
// @Success 200 {object} []string
// @Router /template/tfvars/search [get]
func TemplateTfvarsSearch(c *ctx.GinRequestCtx){
	form := forms.TemplateTfvarsSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsTfVarsSearch(c.ServiceCtx(), &form))
}
// TemplateVariableSearch 查询云模板TF参数
// @Summary 查询云模板TF参数
// @Description 查询云模板TF参数
// @Tags 云模板
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param repoId query string true "repoId"
// @Param repoBranch query string true "repo branch"
// @Param vcsId query int true "vcsId"
// @Success 200 {object} services.TemplateVariable
// @Router /template/variable/search [get]
func TemplateVariableSearch(c *ctx.GinRequestCtx){
	form := forms.TemplateVariableSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsVariableSearch(c.ServiceCtx(), &form))
}


// TemplatePlaybookSearch playbook列表查询
// @Summary playbook列表查询
// @Description playbook列表查询
// @Tags 云模板
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param repoId query string true "repoId"
// @Param repoBranch query string true "repo branch"
// @Param vcsId query int true "vcsId"
// @Success 200 {object} services.TemplateVariable
// @Router /template/playbook/search [get]
func TemplatePlaybookSearch(c *ctx.GinRequestCtx){
	form := forms.TemplatePlaybookSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsPlaybookSearch(c.ServiceCtx(), &form))
}


