// Copyright 2021 CloudJ Company Limited. All rights reserved.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Vcs struct {
	ctrl.GinController
}

// Create 创建vcs仓库
// @Tags Vcs仓库
// @Summary 创建vcs仓库
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form formData forms.CreateVcsForm true "parameter"
// @Router /vcs [post]
// @Success 200 {object} ctx.JSONResult
func (Vcs) Create(c *ctx.GinRequest) {
	form := &forms.CreateVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateVcs(c.Service(), form))
}

// Search 查询vcs仓库
// @Tags Vcs仓库
// @Summary 查询vcs仓库
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.SearchVcsForm true "parameter"
// @Router /vcs [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Vcs}}
func (Vcs) Search(c *ctx.GinRequest) {
	form := &forms.SearchVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchVcs(c.Service(), form))
}

// Update 更新vcs仓库
// @Tags Vcs仓库
// @Summary 更新vcs仓库
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form formData forms.UpdateVcsForm true "parameter"
// @Param vcsId path string true "Vcs仓库ID"
// @Router /vcs/{vcsId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.Vcs}
func (Vcs) Update(c *ctx.GinRequest) {
	form := &forms.UpdateVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateVcs(c.Service(), form))
}

// Delete 删除Vcs 仓库
// @Tags Vcs仓库
// @Summary 删除vcs仓库
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form formData forms.DeleteVcsForm true "parameter"
// @Param vcsId path string true "vcs仓库Id"
// @Router /vcs/{vcsId} [delete]
// @Success 200 {object} ctx.JSONResult
func (Vcs) Delete(c *ctx.GinRequest) {
	form := &forms.DeleteVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteVcs(c.Service(), form))
}

func ListEnableVcs(c *ctx.GinRequest) {
	c.JSONResult(apps.ListEnableVcs(c.Service()))
}

// ListRepos 列出Vcs地址下所有的代码仓库
// @Tags Vcs仓库
// @Summary 列出vcs地址下所有代码仓库
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param vcsId path string true "Vcs仓库ID"
// @Param form query forms.GetGitProjectsForm true "parameter"
// @Router /vcs/{vcsId}/repo [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]vcsrv.Projects}}
func (Vcs) ListRepos(c *ctx.GinRequest) {
	form := forms.GetGitProjectsForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ListRepos(c.Service(), &form))
}

// ListBranches 列出代码仓库下所有分支
// @Tags Vcs仓库
// @Summary 列出代码仓库下所有分支
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param vcsId path string true "Vcs仓库ID"
// @Param form query forms.GetGitRevisionForm true "parameter"
// @Router /vcs/{vcsId}/branch [get]
// @Success 200 {object} ctx.JSONResult{result=[]apps.Revision}
func (Vcs) ListBranches(c *ctx.GinRequest) {
	form := forms.GetGitRevisionForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ListRepoBranches(c.Service(), &form))
}

// ListTags 列出代码仓库下tag
// @Tags Vcs仓库
// @Summary 列出代码仓库下所有分支
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param vcsId path string true "Vcs仓库ID"
// @Param form query forms.GetGitRevisionForm true "parameter"
// @Router /vcs/{vcsId}/tag [get]
// @Success 200 {object} ctx.JSONResult{result=[]apps.Revision}
func (Vcs) ListTags(c *ctx.GinRequest) {
	form := forms.GetGitRevisionForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ListRepoTags(c.Service(), &form))
}

// GetReadmeContent 列出代码仓库下Readme 文件内容
// @Tags Vcs仓库
// @Summary 列出代码仓库下 Readme 文件内容
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param vcsId path string true "Vcs仓库ID"
// @Param form query forms.GetReadmeForm true "parameter"
// @Router /vcs/{vcsId}/readme [get]
// @Success 200 {object} ctx.JSONResult{result=string}
func (Vcs) GetReadmeContent(c *ctx.GinRequest) {
	form := forms.GetReadmeForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.GetReadme(c.Service(), &form))
}

// SearchVcsFileContent 查询代码仓库下文件内容
// @Tags Vcs仓库
// @Summary 查询代码仓库下文件内容
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param vcsId path string true "vcs仓库ID"
// @Param form query forms.SearchVcsFileForm true "parameter"
// @Router /vcs/{vcsId}/file [get]
// @Success 200 {object} ctx.JSONResult{result=string}
func (Vcs) SearchVcsFileContent(c *ctx.GinRequest) {
	form := forms.SearchVcsFileForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchVcsFile(c.Service(), &form))
}
