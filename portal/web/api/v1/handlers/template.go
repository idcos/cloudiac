// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/common"
	"cloudiac/portal/apps"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"cloudiac/utils/logs"
	"encoding/json"
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
	c.JSONResult(apps.VcsFileSearch(c.Service(), &form))
}

// TemplateVariableSearch 查询云模板TF参数
// @Tags 云模板
// @Summary 云模板参数接口
// @Accept application/x-www-form-urlencoded
// @Param IaC-Org-Id header string true "组织ID"
// @Security AuthToken
// @Param form query forms.TemplateVariableSearchForm true "parameter"
// @Param vcsId path string true "vcs地址iD"
// @Router /vcs/{vcsId}/repos/variables [get]
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
// @Tags 云模板
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

// TemplateExport 云模板导出
// @Tags 云模板
// @Summary 云模板导出接口
// @Accept application/json, application/x-www-form-urlencoded
// @Param IaC-Org-Id header string true "组织ID"
// @Security AuthToken
// @Param form query apps.TplExportForm true "parameter"
// @Router /templates/export [get]
// @Success 200 {object} ctx.JSONResult{result=services.TplExportedData}
func TemplateExport(c *ctx.GinRequest) {
	form := apps.TplExportForm{}
	if err := c.Bind(&form); err != nil {
		return
	}

	resp, err := apps.TemplateExport(c.Service(), &form)
	if err != nil {
		c.JSONError(err)
	}

	if !form.Download {
		c.JSONResult(resp, err)
		return
	} else {
		data, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			logs.Get().Warnf("json.Marshal: %v", err)
			c.JSONError(e.New(e.JSONParseError))
		}
		c.FileDownloadResponse(data, "cloudiac-templates.json", "")
	}
}

// TemplateImport 云模板导入
// @Tags 云模板
// @Summary 云模板导入接口
// @Accept application/json, multipart/form-data
// @Param IaC-Org-Id header string true "组织ID"
// @Security AuthToken
// @Param form body apps.TplImportForm true "parameter"
// @Param data body services.TplExportedData false "待导入数据(与 file 参数二选一)"
// @Param file formData file false "待导入文件(与 data 参数二选一)"
// @Router /templates/import [post]
// @Success 200 {object} ctx.JSONResult{result=services.TplImportResult}
func TemplateImport(c *ctx.GinRequest) {
	form := apps.TplImportForm{}
	if err := c.Bind(&form); err != nil {
		return
	}

	if form.File != nil {
		file, err := form.File.Open()
		if err != nil {
			c.JSONError(e.New(e.BadParam, err))
			return
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&form.Data); err != nil {
			c.JSONError(e.New(e.BadParam, err))
			return
		}
	}
	c.JSONResult(apps.TemplateImport(c.Service(), &form))
}
