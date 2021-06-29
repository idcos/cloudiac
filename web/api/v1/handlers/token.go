package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type Token struct {
	ctrl.BaseController
}

// Create 创建ApiToken
// @Summary 创建ApiToken
// @Description 创建ApiToken
// @Tags ApiToken
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.CreateTokenForm true "ApiToken信息"
// @Success 200 {object} models.Token
// @Router /token/create [post]
func (Token) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateToken(c.ServiceCtx(), form))
}

// Search 查询ApiToken列表
// @Summary 查询ApiToken列表
// @Description 查询ApiToken列表
// @Tags ApiToken
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param q query string false "模糊搜索"
// @Param status query string false "ApiToken状态"
// @Success 200 {object} models.Token
// @Router /token/search [get]
func (Token) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchToken(c.ServiceCtx(), form))
}

// Update 修改ApiToken信息
// @Summary 修改ApiToken信息
// @Description 修改ApiToken信息
// @Tags ApiToken
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.UpdateTokenForm true "ApiToken信息"
// @Success 200 {object} models.Token
// @Router /token/update [put]
func (Token) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateToken(c.ServiceCtx(), form))
}

// Delete 删除ApiToken账号
// @Summary 删除ApiToken账号
// @Description 删除ApiToken账号
// @Tags ApiToken
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.DeleteTokenForm true "DeleteTokenForm信息"
// @Success 200
// @Router /token/delete [delete]
func (Token) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteToken(c.ServiceCtx(), form))
}
