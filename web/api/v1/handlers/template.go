package handlers

import (
	"cloudiac/apps"
	"cloudiac/consts/e"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type Template struct {
	ctrl.BaseController
}

// Create 创建云模板
// @Summary 创建云模板
// @Description 创建云模板
// @Tags 云模板
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.CreateTemplateForm true "云模板信息"
// @Success 200 {object} models.Template
// @Router /template/create [post]
func (Template) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateTemplateForm{}
	if err := c.Bind(form); err != nil {
		return
	}
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
// @Param status query string false "模板状态"
// @Param taskStatus query string false "作业运行状态"
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
// @Success 200 {object} models.Template
// @Router /template/update [put]
func (Template) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateTemplate(c.ServiceCtx(), &form))
}


func (Template) Delete(c *ctx.GinRequestCtx) {
	//form := forms.DeleteUserForm{}
	//if err := c.Bind(&form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.DeleteUser(c.ServiceCtx(), &form))
	c.JSONError(e.New(e.NotImplement))
}

// Detail 查询云模板详情
// @Summary 查询云模板详情
// @Description 查询云模板详情
// @Tags 云模板
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param id query int true "云模板id"
// @Success 200 {object} models.Template
// @Router /template/detail [get]
func (Template) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DetailTemplate(c.ServiceCtx(), &form))
}

// Overview 查询云模板概览
// @Summary 查询云模板概览
// @Description 查询云模板概览
// @Tags 云模板
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param id query int true "云模板id"
// @Success 200 {object} apps.OverviewTemplateResp
// @Router /template/overview [get]
func (Template) Overview(c *ctx.GinRequestCtx) {
	form := forms.OverviewTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.OverviewTemplate(c.ServiceCtx(), &form))
}
