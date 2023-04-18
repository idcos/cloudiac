// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type ResourceAccount struct {
	ctrl.GinController
}

// Search 搜索资源账号
// @Tags 资源账号
// @Summary 查询资源账号
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.SearchResourceAccountForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.SearchResourceAccountResp}}
// @Router /resource/account [get]
func (ResourceAccount) Search(c *ctx.GinRequest) {
	form := &forms.SearchResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchResourceAccount(c.Service(), form))
}

// Create 创建资源账号
// @Tags 资源账号
// @Summary 创建资源账号
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form body forms.CreateResourceAccountForm true "parameter"
// @router /resource/account [post]
// @Success 200 {object} ctx.JSONResult{result=models.ResourceAccount}
func (ResourceAccount) Create(c *ctx.GinRequest) {
	form := &forms.CreateResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateResourceAccount(c.Service(), form))
}

// Delete 删除资源账号
// @Tags 资源账号
// @Summary 删除资源账号
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param id path string true "资源账号id"
// @router /resource/account/{id} [delete]
// @Success 200 {object} ctx.JSONResult
func (ResourceAccount) Delete(c *ctx.GinRequest) {
	form := &forms.DeleteResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteResourceAccount(c.Service(), form))
}

// Update 修改资源账号
// @Tags 资源账号
// @Summary 修改资源账号
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param resourceAccountId path string true "资源账号id"
// @Param form body forms.UpdateResourceAccountForm true "parameter"
// @router /resource/account/{resourceAccountId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.ResourceAccount}
func (ResourceAccount) Update(c *ctx.GinRequest) {
	form := &forms.UpdateResourceAccountForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateResourceAccount(c.Service(), form))
}
