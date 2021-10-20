// Copyright 2021 CloudJ Company Limited. All rights reserved.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type VariableGroup struct {
	ctrl.GinController
}


// Search 查询变量组
// @Tags 变量组
// @Summary 查询变量组
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.SearchVariableGroupForm true "parameter"
// @router /var_groups [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.VariableGroup}}
func (VariableGroup) Search(c *ctx.GinRequest) {
	form := forms.SearchVariableGroupForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchVariableGroup(c.Service(), &form))
}

// Create 创建变量组
// @Tags 变量组
// @Summary 创建变量组
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.CreateVariableGroupForm true "parameter"
// @router /var_groups [post]
// @Success 200 {object} ctx.JSONResult{result=models.VariableGroup}
func (VariableGroup) Create(c *ctx.GinRequest) {
	form := forms.CreateVariableGroupForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.CreateVariableGroup(c.Service(), &form))
}

// Update 修改变量组
// @Tags 变量组
// @Summary 修改变量组
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.UpdateVariableGroupForm true "parameter"
// @router /var_groups/{group_id} [put]
// @Success 200 {object} ctx.JSONResult{}
func (VariableGroup) Update(c *ctx.GinRequest) {
	form := forms.UpdateVariableGroupForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateVariableGroup(c.Service(), &form))
}

// Delete 删除变量组
// @Tags 变量组
// @Summary 删除变量组
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.DeleteVariableGroupForm true "parameter"
// @router /var_groups/{group_id} [delete]
// @Success 200 {object} ctx.JSONResult{}
func (VariableGroup) Delete(c *ctx.GinRequest) {
	form := forms.DeleteVariableGroupForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteVariableGroup(c.Service(), &form))
}

// Detail 查询变量组详情
// @Tags 变量组
// @Summary 查询变量组详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.DeleteVariableGroupForm true "parameter"
// @router /var_groups/{group_id} [get]
// @Success 200 {object} ctx.JSONResult{result=models.VariableGroup}
func (VariableGroup) Detail(c *ctx.GinRequest) {
	form := forms.DetailVariableGroupForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DetailVariableGroup(c.Service(), &form))
}
