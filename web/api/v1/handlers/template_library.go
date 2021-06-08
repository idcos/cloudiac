package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type TemplateLibrary struct {
	ctrl.BaseController
}

func TemplateLibraryHandler(c *ctx.GinRequestCtx) {
	c.JSONResult("",nil)

}

// Create
// @Tags 云模板库
// @Description 通过云模板库创建云模板接口
// @Accept application/json
// @Param id formData int false "云模板库id"
// @router /api/v1/template/library/create [post]
// @Success 200 {object} models.TemplateLibrary
func (TemplateLibrary) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateTemplateLibraryForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateTemplateLibrary(c.ServiceCtx(), form))
}
// Search
// @Accept application/json
// @Tags 云模板库
// @Description 通过云模板库列表查询接口
// @router /api/v1/template/library/search [get]
// @Success 200 {object} models.TemplateLibrary
func (TemplateLibrary) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchTemplateLibraryForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchTemplateLibrary(c.ServiceCtx(), form))
}
