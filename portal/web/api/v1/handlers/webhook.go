package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type AccessToken struct {
	ctrl.BaseController
}

// Create 创建webhook
// @Tags 触发器
// @Summary 创建触发器接口
// @Accept application/json
// @Param tplGuid formData string false "云模版guid"
// @Param action formData string false "动作"
// @router /webhook/create [post]
// @Success 200 {object} models.TemplateAccessToken
func (AccessToken) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateAccessTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateAccessToken(c.ServiceCtx(), form))
}

// Search 查询webhook
// @Tags 触发器
// @Summary 查询触发器接口
// @Accept application/json
// @Param tplGuid formData string false "云模版guid"
// @router /webhook/search [get]
// @Success 200 {object} models.TemplateAccessToken
func (AccessToken) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchAccessTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchAccessToken(c.ServiceCtx(), &form))
}

// Update 修改webhook
// @Tags 触发器
// @Summary 修改触发器接口
// @Accept application/json
// @Param id formData int false "触发器id"
// @Param action formData string false "动作"
// @router /webhook/update [put]
// @Success 200 {object} models.TemplateAccessToken
func (AccessToken) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateAccessTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateAccessToken(c.ServiceCtx(), &form))
}

// Delete 删除webhook
// @Tags 触发器
// @Summary 删除触发器接口
// @Accept application/json
// @Param id formData int false "触发器id"
// @router /webhook/delete [delete]
// @Success 200 {object} models.TemplateAccessToken
func (AccessToken) Delete(c *ctx.GinRequestCtx) {
	form := forms.DeleteAccessTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteAccessToken(c.ServiceCtx(), &form))
}

// Detail webhook详情
// @Tags 触发器
// @Summary 触发器详情接口
// @Accept application/json
// @Param id formData int false "触发器id"
// @router /webhook/detail [get]
// @Success 200 {object} models.TemplateAccessToken
func (AccessToken) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailAccessTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DetailAccessToken(c.ServiceCtx(), &form))
}
