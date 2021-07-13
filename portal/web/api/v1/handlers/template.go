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
// @Tags 云模版
// @Summary 创建云模版
// @Accept multipart/form-data
// @Accept json
// @Security AuthToken
// @Produce json
// @Param Iac-Org-Id header string true "组织ID"
// @Param form formData forms.CreateTemplateForm true "parameter"
// @Router /template/create [post]
// @Success 200 {object} ctx.JSONResult{result=models.Template}
func (Template) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateTemplateForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateTemplate(c.ServiceCtx(), form))
}


// Search 查询云模板列表
// @Tags 云模版
// @Summary 查询云模板列表
// @Accept multipart/x-www-form-urlencoded
// @Security AuthToken
// @Description 查询云模板列表
// @Tags 云模板
// @Accept  json
// @Produce  json
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-project-Id header string false "项目ID"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]apps.SearchTemplateResp}}
// @Router /template/search [get]
func (Template) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchTemplate(c.ServiceCtx(), &form))
}

// Update 修改云模板信息
// @Tags 云模版
// @Summary 修改云模板信息
// @Description 修改云模板信息
// @Security AuthToken
// @Accept  json
// @Produce  json
// @Param data body forms.UpdateTemplateForm true "云模板信息"
// @Success 200 {object} ctx.JSONResult{result=models.Template}
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
// @Tags 云模版
// @Description 需要组织管理员权限。
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param form formData forms.DeleteUserForm true "parameter"
// @Param templateId path string true "模版ID"
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
// @Tags 云模版
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param form formData forms.DetailTemplateForm true "parameter"
// @Router /template/{templateId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.Template}
func (Template) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.TemplateDetail(c.ServiceCtx(), &form))
}
