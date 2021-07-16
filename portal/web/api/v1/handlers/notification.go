package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Notification struct {
	ctrl.BaseController
}

// Search 查询通知
// @Summary 查询通知
// @Description 查询通知
// @Tags 通知
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]services.NotificationResp}}
// @Router /notifications [get]
func (Notification) Search(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.SearchNotification(c.ServiceCtx()))
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
// @Param form formData forms.CreateNotificationCfgForm true "parameter"
// @Router /notifications [post]
// @Success 200 {object} ctx.JSONResult{result=models.NotificationCfg}
func (Notification) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateNotificationCfgForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateNotificationCfg(c.ServiceCtx(), form))
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
// @Param data body forms.DeleteNotificationCfgForm true "DeleteTokenForm信息"
// @Success 200
// @Router /notifications/{id} [delete]
func (Notification) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteNotificationCfgForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteNotificationCfg(c.ServiceCtx(), form.Id))
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
// @Param data body forms.UpdateNotificationCfgForm true "ApiToken信息"
// @Success 200
// @Router /notifications/{id} [put]
func (Notification) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateNotificationCfgForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateNotificationCfg(c.ServiceCtx(), form))
}
