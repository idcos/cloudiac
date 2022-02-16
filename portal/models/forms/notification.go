// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type UpdateNotificationForm struct {
	BaseForm
	Id        models.Id `uri:"id" form:"notificationId" json:"notificationId" binding:"required"`
	Name      string    `json:"name" form:"name" `
	Type      string    `form:"type" json:"type" binding:"required"`
	Secret    string    `json:"secret" form:"secret"`
	Url       string    `json:"url" form:"url"`
	UserIds   []string  `form:"userIds" json:"userIds"`
	EventType []string  `form:"eventType" json:"eventType" binding:"required"` //enum('task.failed', 'task.complete', 'task.approving', 'task.running', "task.crondrift")
}

type CreateNotificationForm struct {
	BaseForm
	Name      string   `json:"name" form:"name" `
	Type      string   `form:"type" json:"type" binding:"required"`
	Secret    string   `json:"secret" form:"secret"`
	Url       string   `json:"url" form:"url"`
	UserIds   []string `form:"userIds" json:"userIds"`
	EventType []string `form:"eventType" json:"eventType" binding:"required"` //enum('task.failed', 'task.complete', 'task.approving', 'task.running', "task.crondrift")
}

type DeleteNotificationForm struct {
	BaseForm
	Id models.Id `uri:"id" form:"id" json:"id" binding:"required"`
}

type DetailNotificationForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"`
}

type SearchNotificationForm struct {
	PageForm
}
