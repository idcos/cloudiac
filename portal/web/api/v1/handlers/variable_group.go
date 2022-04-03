// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

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
// @Param Iac-Project-Id header string false "项目ID"
// @Param form query forms.SearchVariableGroupForm true "parameter"
// @router /var_groups [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.SearchVariableGroupResp}}
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
// @Param json body forms.CreateVariableGroupForm true "parameter"
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
// @Param group_id path string true "变量组id"
// @Param json body forms.UpdateVariableGroupForm true "parameter"
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
// @Param group_id path string true "变量组id"
// @router /var_groups/{group_id} [delete]
// @Success 200 {object} ctx.JSONResult
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
// @Param group_id path string true "变量组id"
// @router /var_groups/{group_id} [get]
// @Success 200 {object} ctx.JSONResult{result=models.VariableGroup}
func (VariableGroup) Detail(c *ctx.GinRequest) {
	form := forms.DetailVariableGroupForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DetailVariableGroup(c.Service(), &form))
}

// SearchRelationship 查询变量组与实例的关系
// @Tags 变量组
// @Summary 查询变量组与实例的关系
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.SearchRelationshipForm true "parameter"
// @router /var_groups/relationship [get]
// @Success 200 {object} ctx.JSONResult{result=[]services.VarGroupRel}
func (VariableGroup) SearchRelationship(c *ctx.GinRequest) {
	form := forms.SearchRelationshipForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchRelationship(c.Service(), &form))
}

// BatchUpdateRelationship 绑定变量组与实例的关系
// @Tags 变量组
// @Summary 绑定变量组与实例的关系
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param json body forms.BatchUpdateRelationshipForm true "parameter"
// @router /var_groups/relationship/batch [put]
// @Success 200 {object} ctx.JSONResult
func (VariableGroup) BatchUpdateRelationship(c *ctx.GinRequest) {
	form := forms.BatchUpdateRelationshipForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.BatchUpdateRelationship(c.Service(), &form))
}

// DeleteRelationship 删除变量组与实例的关系
// @Tags 变量组
// @Summary 删除变量组与实例的关系
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param varGroupId path string true "变量组id"
// @router /var_groups/relationship/{varGroupId} [delete]
// @Success 200 {object} ctx.JSONResult{}
func (VariableGroup) DeleteRelationship(c *ctx.GinRequest) {
	form := forms.DeleteRelationshipForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteRelationship(c.Service(), &form))
}

// SearchRelationshipAll 查询变量组与实例的关系(不进行继承处理展示所有的变量)
// @Tags 变量组
// @Summary 查询变量组与实例的关系(不进行继承处理展示所有的变量)
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.SearchRelationshipForm true "parameter"
// @router /var_groups/relationship/all [get]
// @Success 200 {object} ctx.JSONResult{result=[]services.VarGroupRel}
func (VariableGroup) SearchRelationshipAll(c *ctx.GinRequest) {
	form := forms.SearchRelationshipForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchRelationshipAll(c.Service(), &form))
}
