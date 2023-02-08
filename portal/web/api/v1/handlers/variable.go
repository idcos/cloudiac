// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
)

type Variable struct {
	ctrl.GinController
}

// BatchUpdate 批量修改变量(己弃用)
// @Tags 变量
// @Summary 批量修改变量
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form body forms.BatchUpdateVariableForm true "parameter"
// @router /variables/batch [put]
// @Success 200 {object} ctx.JSONResult
func (Variable) BatchUpdate(c *ctx.GinRequest) {
	form := forms.BatchUpdateVariableForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.BatchUpdate(c.Service(), &form))
}

// UpdateObjectVars 更新实例的变量
// @Tags 变量
// @Summary 更新实例的变量
// @Description 该接口全量更新实例的变量列表，缺少的变量会被创建，多余的变量会被删除
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param scope path string true "变量的作用域(即实例类型)" enums(org,template,project,env)
// @Param objectId path string true "变量关联的实例 id"
// @Param form body forms.UpdateObjectVarsForm true "parameter"
// @router /variables/scope/{scope}/{objectId} [put]
// @Success 200 {object} ctx.JSONResult{result=[]models.Variable}
func (Variable) UpdateObjectVars(c *ctx.GinRequest) {
	form := forms.UpdateObjectVarsForm{}
	form.Scope = c.Param("scope")
	form.ObjectId = models.Id(c.Param("id"))
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateObjectVars(c.Service(), &form))
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
// @Success 200 {object} ctx.JSONResult{result=[]resps.VariableResp}
func (Variable) Search(c *ctx.GinRequest) {
	form := forms.SearchVariableForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchVariable(c.Service(), &form))
}

// SearchSampleVariable 查询变量
// @Tags 变量
// @Summary 查询变量
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.SearchVariableForm true "parameter"
// @router /variables/sample [get]
// @Success 200 {object} ctx.JSONResult{result=[]models.VariableBody}
func (Variable) SearchSampleVariable(c *ctx.GinRequest) {
	form := forms.SearchVariableForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchSampleVariable(c.Service(), &form))
}
