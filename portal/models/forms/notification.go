// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package forms

import "cloudiac/portal/models"

type UpdateNotificationForm struct {
	BaseForm
	Id        models.Id `uri:"id" form:"notificationId" json:"notificationId" binding:"required,startswith=notif-,max=32" swaggerignore:"true"`
	Name      string    `json:"name" form:"name" binding:"omitempty,gte=2,lte=255"`
	Type      string    `form:"type" json:"type" binding:"omitempty,oneof=email webhook wechat slack dingtalk"`
	Secret    string    `json:"secret" form:"secret" binding:"max=255"`
	Url       string    `json:"url" form:"url" binding:"omitempty,url,max=255"` //url格式
	UserIds   []string  `form:"userIds" json:"userIds" binding:"omitempty,dive,required,startswith=u-,max=32"`
	EventType []string  `form:"eventType" json:"eventType" binding:"omitempty,dive,required,startswith=task."` //enum('task.failed', 'task.complete', 'task.approving', 'task.running', "task.crondrift")
}

type CreateNotificationForm struct {
	BaseForm
	Name      string   `json:"name" form:"name" binding:"required,gte=2,lte=255"`
	Type      string   `form:"type" json:"type" binding:"required,oneof=email webhook wechat slack dingtalk"`
	Secret    string   `json:"secret" form:"secret" binding:"max=255"`
	Url       string   `json:"url" form:"url" binding:"omitempty,url,max=255"`
	UserIds   []string `form:"userIds" json:"userIds" binding:"omitempty,dive,required,startswith=u-,max=32"`
	EventType []string `form:"eventType" json:"eventType" binding:"omitempty,dive,required,startswith=task."` //enum('task.failed', 'task.complete', 'task.approving', 'task.running', "task.crondrift")
}

type DeleteNotificationForm struct {
	BaseForm
	Id models.Id `uri:"id" form:"id" json:"id" binding:"required,startswith=notif-,max=32" swaggerignore:"true"`
}

type DetailNotificationForm struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" swaggerignore:"true" binding:"required,startswith=notif-,max=32"`
}

type SearchNotificationForm struct {
	PageForm
}
