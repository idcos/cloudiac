package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type AccessToken struct {
	ctrl.BaseController
}

// Create 创建触发器
// @Summary 创建触发器
// @Description 创建触发器接口
// @Tags 触发器
// @Accept   json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.CreateAccessTokenForm true "触发器信息"
// @Success 200 {object} models.TemplateAccessToken
// @Router  /webhook/create [post]
func (AccessToken) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateAccessTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateAccessToken(c.ServiceCtx(), form))
}

// Search 查询触发器
// @Summary 查询触发器
// @Description 查询触发器接口
// @Tags 触发器
// @Accept   json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param tplGuid query string true "云模版guid"
// @Success 200 {object} models.TemplateAccessToken
// @Router  /webhook/search [get]
func (AccessToken) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchAccessTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchAccessToken(c.ServiceCtx(), &form))
}

// Update 修改触发器
// @Summary 修改触发器
// @Description 修改触发器接口
// @Tags 触发器
// @Accept   json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.UpdateAccessTokenForm true "触发器信息"
// @Success 200 {object} models.TemplateAccessToken
// @Router  /webhook/update [put]
func (AccessToken) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateAccessTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateAccessToken(c.ServiceCtx(), &form))
}

// Delete 删除触发器
// @Summary 删除触发器
// @Description 删除触发器接口
// @Tags 触发器
// @Accept   json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.DeleteAccessTokenForm true "触发器信息"
// @Success 200 {object} models.TemplateAccessToken
// @Router  /webhook/delete [delete]
func (AccessToken) Delete(c *ctx.GinRequestCtx) {
	form := forms.DeleteAccessTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteAccessToken(c.ServiceCtx(), &form))
}

// Detail 触发器详情
// @Summary 触发器详情
// @Description 触发器详情接口
// @Tags 触发器
// @Accept   json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param id query string true "触发器id"
// @Success 200 {object} models.TemplateAccessToken
// @Router  /webhook/detail [get]
func (AccessToken) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailAccessTokenForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DetailAccessToken(c.ServiceCtx(), &form))
}
