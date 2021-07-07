package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Template struct {
	ctrl.BaseController
}
// 创建云模版
// @Summary 创建云模版
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Param name formData string true "模版名称"
// @Param vcsId formData int true "vcs仓库"
// @Param tplType formData string true "云模版类型"
// @Param orgId formData string true "组织id"
// @Param description formData string false "云模版描述信息"
// @Param repoId formData string true "云模版代码仓库id"
// @Param repoAddr formData string true "云模版代码仓库地址"
// @Param repoRevision formData string false "云模版仓库分支,默认值为master"
// @Param workdir formData string false "工作路径"
// @Param playbook formData string false "ansbile playbook文件路径"
// @Param status formData string false "云模版状态，有enable, disable两个可选值，默认值为enable"
// @Param creatorId formData string true "创建用户ID"
// @Param runnerId formData string true "runnerId"
// @Param varFile formData string false "tfvars 文件路径"
// @Router /template/create [post]
// @Success 200 {object} ctx.JSONResult{result=models.Template}
// TODO 少vars 和 tfvars文件，这两个不在表里面
func (Template) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateTemplateForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	// TODO 缺少关联项目的
	c.JSONResult(apps.CreateTemplate(c.ServiceCtx(), form))
}


// Search 查询云模板列表
// @Summary 查询云模板列表
// @Description 查询云模板列表
// @Tags 云模板
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param q query string false "模糊搜索"
// @Success 200 {object} apps.SearchTemplateResp
// @Router /template/search [get]
func (Template) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchTemplate(c.ServiceCtx(), &form))
}



// Update 修改云模板信息
// @Summary 修改云模板信息
// @Description 修改云模板信息
// @Tags 云模板
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.UpdateTemplateForm true "云模板信息"
// @Success 200 {object} ctx.JSONResult={models.Template}
// @Router /template/update [put]
func (Template) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateTemplate(c.ServiceCtx(), &form))

}

// 删除活跃云模版
// @Summary 删除云模版
// @Description 需要组织管理员权限。
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param form formData forms.DeleteUserForm true "parameter"
// @Router /template/{templateId} [delete]
// @Success 200 {object} ctx.JSONResult
func (Template) Delete(c *ctx.GinRequestCtx) {
	form := forms.DeleteTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DelateTemplate(c.ServiceCtx(), &form))
}

// 模版详情
// @Summary 模版详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param form formData forms.DetailTemplateForm true
// @Router /template/{templateId}
// @Success 200 {object} ctx.JSONResult{result=models.Template}
func (Template) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.TemplateDetail(c.ServiceCtx(), &form))
}
