// Copyright 2021 CloudJ Company Limited. All rights reserved.

package handlers

import (
	"cloudiac/common"
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Template struct {
	ctrl.GinController
}

// Create 创建云模板
// @Tags 云模板
// @Summary 创建云模板
// @Accept multipart/form-data
// @Accept json
// @Security AuthToken
// @Produce json
// @Param IaC-Org-Id header string true "组织ID"
// @Param form body forms.CreateTemplateForm true "parameter"
// @Router /templates [post]
// @Success 200 {object} ctx.JSONResult{result=models.Template}
func (Template) Create(c *ctx.GinRequest) {
	form := &forms.CreateTemplateForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateTemplate(c.Service(), form))
}

// Search 查询云模板列表
// @Tags 云模板
// @Summary 查询云模板列表
// @Accept multipart/x-www-form-urlencoded
// @Security AuthToken
// @Description 查询云模板列表
// @Tags 云模板
// @Accept  json
// @Produce  json
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.SearchTemplateForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]apps.SearchTemplateResp}}
// @Router /templates [get]
func (Template) Search(c *ctx.GinRequest) {
	form := forms.SearchTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchTemplate(c.Service(), &form))
}

// Update 修改云模板信息
// @Tags 云模板
// @Summary 修改云模板信息
// @Description 修改云模板信息
// @Security AuthToken
// @Accept  json
// @Produce  json
// @Param templateId path string true "云模板ID"
// @Param data body forms.UpdateTemplateForm true "云模板信息"
// @Success 200 {object} ctx.JSONResult{result=models.Template}
// @Router /templates/{templateId} [put]
func (Template) Update(c *ctx.GinRequest) {
	form := forms.UpdateTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateTemplate(c.Service(), &form))
}

// Delete 删除活跃云模板
// @Summary 删除云模板
// @Tags 云模板
// @Description 删除云模板,需要组织管理员权限。
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.DeleteTemplateForm true "parameter"
// @Param templateId path string true "云模板ID"
// @Router /templates/{templateId} [delete]
// @Success 200 {object} ctx.JSONResult
func (Template) Delete(c *ctx.GinRequest) {
	form := forms.DeleteTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteTemplate(c.Service(), &form))
}

// Detail 模板详情
// @Summary 模板详情
// @Tags 云模板
// @Description 获取云模板详情。
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param templateId path string true "云模板ID"
// @Param form query forms.DetailTemplateForm true "parameter"
// @Router /templates/{templateId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.Template}
func (Template) Detail(c *ctx.GinRequest) {
	form := forms.DetailTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.TemplateDetail(c.Service(), &form))
}

// TemplateTfvarsSearch 列出代码仓库下包含.tfvars 的所有文件
// @Tags 云模板
// @Summary 列出代码仓库下.tfvars 的所有文件
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param vcsId path string true "vcs地址iD"
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.TemplateTfvarsSearchForm true "parameter"
// @Router /vcs/{vcsId}/repos/tfvars  [get]
// @Success 200 {object} ctx.JSONResult{result=[]vcsrv.VcsIfaceOptions}
func TemplateTfvarsSearch(c *ctx.GinRequest) {
	form := forms.TemplateTfvarsSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsTfVarsSearch(c.Service(), &form))
}

// TemplateVariableSearch 查询云模板TF参数
// @Tags 云模板
// @Summary 云模板参数接口
// @Accept application/x-www-form-urlencoded
// @Param IaC-Org-Id header string true "组织ID"
// @Security AuthToken
// @Param form query forms.TemplateVariableSearchForm true "parameter"
// @Router /templates/variables [get]
// @Success 200 {object} ctx.JSONResult{result=[]services.TemplateVariable}
func TemplateVariableSearch(c *ctx.GinRequest) {
	form := forms.TemplateVariableSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsVariableSearch(c.Service(), &form))
}

// TemplatePlaybookSearch
// @Tags 云模板
// @Summary  playbook列表接口
// @Accept application/x-www-form-urlencoded
// @Param IaC-Org-Id header string true "组织ID"
// @Security AuthToken
// @Param form query forms.TemplatePlaybookSearchForm true "parameter"
// @Param vcsId path string true "vcs地址iD"
// @router /vcs/{vcsId}/repos/playbook [get]
// @Success 200 {object} ctx.JSONResult{result=[]vcsrv.VcsIfaceOptions}
func TemplatePlaybookSearch(c *ctx.GinRequest) {
	form := forms.TemplatePlaybookSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsPlaybookSearch(c.Service(), &form))
}

// TemplateTfVersionSearch
// @Tags 云模板
// @Summary terraform versions tf版本列表接口
// @Accept application/x-www-form-urlencoded
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @router /templates/tfversions [get]
// @Success 200 {object} []string
func TemplateTfVersionSearch(c *ctx.GinRequest) {
	c.JSONResult(common.TerraformVersions, nil)
}

// AutoTemplateTfVersionChoice
// @Tags 云模板
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Summary choice terraform version 根据用户文件设置自动选择tf版本
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.TemplateTfVersionSearchForm true "parameter"
// @router /templates/autotfversion [GET]
// @Success 200 {object} string
func AutoTemplateTfVersionChoice(c *ctx.GinRequest) {
	form := forms.TemplateTfVersionSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.AutoGetTfVersion(c.Service(), &form))
}

// TemplateChecks
// @Tag 云模版
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Summary 创建云模版前检查名称是否重复和工作目录是否正确
// @Param IaC-Org-Id header string true "组织ID"
// @Security AuthToken
// @Param form query forms.TemplateTfVersionSearchForm true "parameter"
// @router /templates/checks [POST]
// @Param form query forms.TemplateChecksForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=apps.TemplateChecksResp}
func TemplateChecks(c *ctx.GinRequest) {
	form := forms.TemplateChecksForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.TemplateChecks(c.Service(), &form))
}
