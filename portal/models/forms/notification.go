// Copyright 2021 CloudJ Company Limited. All rights reserved.

package forms

import "cloudiac/portal/models"

type UpdateNotificationForm struct {
	BaseForm
	Id      models.Id `uri:"id" form:"notificationId" json:"notificationId" binding:"required"`
	Name    string    `json:"name" form:"name" `
	Type    string    `form:"type" json:"type" binding:"required"`
	Secret  string    `json:"secret" form:"secret"`
	Url     string    `json:"url" form:"url"`
	UserIds []string  `form:"userIds" json:"userIds"`
	//EventType        string      `form:"eventType" json:"eventType" binding:"required"`

	EventType []string `form:"eventType" json:"eventType" binding:"required"` //enum('failed', 'complete', 'approving', 'running')
}

type CreateNotificationForm struct {
	BaseForm
	Name      string   `json:"name" form:"name" `
	Type      string   `form:"notificationType" json:"notificationType" binding:"required"`
	Secret    string   `json:"secret" form:"secret"`
	Url       string   `json:"url" form:"url"`
	UserIds   []string `form:"userIds" json:"userIds"`
	EventType []string `form:"eventType" json:"eventType" binding:"required"` //enum('failed', 'complete', 'approving', 'running')

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
