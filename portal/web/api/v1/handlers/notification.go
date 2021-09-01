// Copyright 2021 CloudJ Company Limited. All rights reserved.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Notification struct {
	ctrl.GinController
}

// Search 查询通知
// @Summary 查询通知
// @Description 查询通知
// @Tags 通知
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]services.RespNotification}}
// @Router /notifications [get]
func (Notification) Search(c *ctx.GinRequest) {
	c.JSONResult(apps.SearchNotification(c.Service()))
}

// Create 创建通知
// @Tags 通知
// @Summary 创建通知
// @Description 创建通知
// @Accept multipart/form-data
// @Accept json
// @Security AuthToken
// @Produce json
// @Param IaC-Org-Id header string true "组织ID"
// @Param json body forms.CreateNotificationForm true "parameter"
// @Router /notifications [post]
// @Success 200 {object} ctx.JSONResult{result=models.Notification}
func (Notification) Create(c *ctx.GinRequest) {
	form := &forms.CreateNotificationForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateNotification(c.Service(), form))
}

// Delete 删除通知信息
// @Summary 删除通知信息
// @Description 删除Token账号
// @Tags Token
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param id path string true "通知id"
// @Param data body forms.DeleteNotificationForm true "DeleteTokenForm信息"
// @Success 200
// @Router /notifications/{id} [delete]
func (Notification) Delete(c *ctx.GinRequest) {
	form := &forms.DeleteNotificationForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteNotification(c.Service(), form.Id))
}

// Update 修改通知信息
// @Summary 修改通知信息
// @Description 修改通知信息
// @Tags 通知
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param id path string true "通知id"
// @Param data body forms.UpdateNotificationForm true "ApiToken信息"
// @Success 200
// @Router /notifications/{id} [put]
func (Notification) Update(c *ctx.GinRequest) {
	form := &forms.UpdateNotificationForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateNotification(c.Service(), form))
}

// Detail 查询通知详情
// @Summary 查询通知详情
// @Description 查询通知详情
// @Tags 通知
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Success 200 {object} ctx.JSONResult{result=models.Notification}
// @Router /notifications/{notificationId}  [get]
func (Notification) Detail(c *ctx.GinRequest) {
	form := &forms.DetailNotificationForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailNotification(c.Service(), form))
}
