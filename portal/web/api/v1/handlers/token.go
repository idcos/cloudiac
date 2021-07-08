package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Token struct {
	ctrl.BaseController
}

// Create 创建token
// @Summary 创建token
// @Description 创建token
// @Tags Token
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param data body forms.CreateTokenForm true "ApiToken信息"
// @Success 200 {object} ctx.JSONResult{result=models.Token}
// @Router /token [post]
func (Token) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateToken(c.ServiceCtx(), form))
}

// Search 查询token
// @Summary 查询token
// @Description 查询token
// @Tags Token
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param q query string false "模糊搜索"
// @Param status query string false "ApiToken状态"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Token}}
// @Router /token [get]
func (Token) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchToken(c.ServiceCtx(), form))
}

// Update 修改token信息
// @Summary 修改token信息
// @Description 修改token信息
// @Tags Token
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param data body forms.UpdateTokenForm true "ApiToken信息"
// @Success 200
// @Router /token/{id} [put]
func (Token) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateToken(c.ServiceCtx(), form))
}

// Delete 删除Token账号
// @Summary 删除Token账号
// @Description 删除Token账号
// @Tags Token
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param data body forms.DeleteTokenForm true "DeleteTokenForm信息"
// @Success 200
// @Router /token/{id} [delete]
func (Token) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteToken(c.ServiceCtx(), form))
}
