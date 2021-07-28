package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Variable struct {
	ctrl.GinController
}

// BatchUpdate 批量修改变量
// @Tags 变量
// @Summary 批量修改变量
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form body forms.BatchUpdateVariableForm true "parameter"
// @router /variables/batch [put]
// @Success 200 {object} ctx.JSONResult{result=models.Variable}
func (Variable) BatchUpdate(c *ctx.GinRequest) {
	form := forms.BatchUpdateVariableForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.BatchUpdate(c.Service(), &form))
}

// Search 查询变量
// @Tags 变量
// @Summary 查询变量
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.SearchVariableForm true "parameter"
// @router /variables [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Variable}}
func (Variable) Search(c *ctx.GinRequest) {
	form := forms.SearchVariableForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchVariable(c.Service(), &form))
}
