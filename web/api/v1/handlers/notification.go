package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type Notification struct {
	ctrl.BaseController
}

// Search 查询组织通知列表
// @Summary 查询组织通知列表
// @Description 查询组织通知列表
// @Tags 组织通知
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} services.NotificationResp
// @Router /notification/search [get]
func (Notification) Search(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.ListNotificationCfgs(c.ServiceCtx()))
}

// Create 创建组织通知
// @Summary 创建组织通知
// @Description 创建组织通知
// @Tags 组织通知
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.CreateNotificationCfgForm true "组织通知信息"
// @Success 200 {object} models.NotificationCfg
// @Router /notification/create [post]
func (Notification) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateNotificationCfgForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateNotificationCfg(c.ServiceCtx(), form))
}

// Delete 修改组织通知删除
// @Summary 修改组织通知删除
// @Description 修改组织通知删除
// @Tags 组织通知
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.DeleteNotificationCfgForm true "组织通知信息"
// @Success 200
// @Router /notification/delete [delete]
func (Notification) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteNotificationCfgForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteNotificationCfg(c.ServiceCtx(), form.Id))
}

// Update 修改组织通知信息
// @Summary 修改组织通知信息
// @Description 修改组织通知信息
// @Tags 组织通知
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.UpdateNotificationCfgForm true "组织通知信息"
// @Success 200 {object} models.NotificationCfg
// @Router /notification/update [put]
func (Notification) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateNotificationCfgForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateNotificationCfg(c.ServiceCtx(), form))
}
