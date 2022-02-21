// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers //nolint:dupl

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Key struct {
	ctrl.GinController
}

// Create 创建密钥
// @Summary 创建密钥
// @Tags 密钥
// @Accept multipart/form-data
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param data formData forms.CreateKeyForm true "密钥信息"
// @Router /keys [post]
// @Success 200 {object} ctx.JSONResult{result=models.Key}
func (Key) Create(c *ctx.GinRequest) {
	form := &forms.CreateKeyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateKey(c.Service(), form))
}

// Search 查询密钥
// @Summary 查询密钥
// @Tags 密钥
// @Accept application/x-www-form-urlencoded
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param data query forms.SearchKeyForm true "密钥查询参数"
// @Router /keys [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Key}}
func (Key) Search(c *ctx.GinRequest) {
	form := &forms.SearchKeyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchKey(c.Service(), form))
}

// Update 修改密钥信息
// @Summary 修改密钥信息
// @Tags 密钥
// @Accept multipart/form-data
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param keyId path string true "密钥ID"
// @Param data formData forms.UpdateKeyForm true "密钥信息"
// @Router /keys/{keyId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.Key}
func (Key) Update(c *ctx.GinRequest) {
	form := &forms.UpdateKeyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateKey(c.Service(), form))
}

// Delete 删除密钥
// @Summary 删除密钥
// @Tags 密钥
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param keyId path string true "密钥ID"
// @Router /keys/{keyId} [delete]
// @Success 200
func (Key) Delete(c *ctx.GinRequest) {
	form := &forms.DeleteKeyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteKey(c.Service(), form))
}

// Detail 密钥详情
// @Summary 密钥详情
// @Tags 密钥
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param keyId path string true "密钥ID"
// @Router /keys/{keyId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.Key}
func (Key) Detail(c *ctx.GinRequest) {
	form := &forms.DetailKeyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailKey(c.Service(), form))
}
