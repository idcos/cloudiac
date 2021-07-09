package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Variable struct {
	ctrl.BaseController
}

// Create 创建变量
// @Tags 变量
// @Summary 创建变量
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form body forms.CreateVariableForm true "parameter"
// @router /variables [post]
// @Success 200 {object} ctx.JSONResult{result=models.Variable}
func (Variable) Create(c *ctx.GinRequestCtx) {
	form := forms.CreateVariableForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.CreateVariable(c.ServiceCtx(), &form))
}

// Search 查询变量
// @Tags 变量
// @Summary 查询变量
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param form body forms.SearchVariableForm true "parameter"
// @router /variables [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Variable}}
func (Variable) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchVariableForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchVariable(c.ServiceCtx(), &form))
}
