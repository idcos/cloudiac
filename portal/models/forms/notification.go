// Copyright 2021 CloudJ Company Limited. All rights reserved.

package forms

import "cloudiac/portal/models"

type CfgInfo struct {
	EmailAddress string    `form:"emailAddress" json:"emailAddress"`
	UserId       models.Id `form:"userId" json:"userId"`
	WebUrl       string    `form:"webUrl" json:"webUrl"`
	UserName     string    `form:"userName" json:"userName"`
}

type UpdateNotificationForm struct {
	PageForm
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
	PageForm
	Name      string   `json:"name" form:"name" `
	Type      string   `form:"notificationType" json:"notificationType" binding:"required"`
	Secret    string   `json:"secret" form:"secret"`
	Url       string   `json:"url" form:"url"`
	UserIds   []string `form:"userIds" json:"userIds"`
	EventType []string `form:"eventType" json:"eventType" binding:"required"` //enum('failed', 'complete', 'approving', 'running')

}

type DeleteNotificationForm struct {
	PageForm
	Id models.Id `uri:"id" form:"id" json:"id" binding:"required"`
}

type DetailNotificationForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true"`
}
