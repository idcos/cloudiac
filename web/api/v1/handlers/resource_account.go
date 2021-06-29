package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type ResourceAccount struct {
	ctrl.BaseController
}

// Search 查询资源账号列表
// @Summary 查询资源账号列表
// @Description 查询资源账号列表
// @Tags 资源账号
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param q query string false "模糊搜索"
// @Param status query string false "资源账号状态"
// @Success 200 {object} apps.searchResourceAccountResp
// @Router /resource/account/search [get]
func (ResourceAccount) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchResourceAccount(c.ServiceCtx(), form))
}

// Create 创建资源账号
// @Summary 创建资源账号
// @Description 创建资源账号
// @Tags 资源账号
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.CreateResourceAccountForm true "资源账号信息"
// @Success 200 {object} models.ResourceAccount
// @Router /resource/account/create [post]
func (ResourceAccount) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateResourceAccount(c.ServiceCtx(), form))
}

// Delete 删除资源账号
// @Summary 删除资源账号
// @Description 删除资源账号
// @Tags 资源账号
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.DeleteResourceAccountForm true "资源账号信息"
// @Success 200
// @Router /resource/account/delete [delete]
func (ResourceAccount) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteResourceAccount(c.ServiceCtx(), form))
}

// Update 修改资源账号信息
// @Summary 修改资源账号信息
// @Description 修改资源账号信息
// @Tags 资源账号
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.UpdateResourceAccountForm true "资源账号信息"
// @Success 200 {object} models.ResourceAccount
// @Router /resource/account/update [put]
func (ResourceAccount) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateResourceAccount(c.ServiceCtx(), form))
}
