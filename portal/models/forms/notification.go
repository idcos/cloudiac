package forms

import "cloudiac/portal/models"

type CfgInfo struct {
	EmailAddress string    `form:"emailAddress" json:"emailAddress"`
	UserId       models.Id `form:"userId" json:"userId"`
	WebUrl       string    `form:"webUrl" json:"webUrl"`
	UserName     string    `form:"userName" json:"userName"`
}

type UpdateNotificationCfgForm struct {
	PageForm
	NotificationId   models.Id `form:"notificationId" json:"notificationId" binding:"required"`
	NotificationType string    `form:"notificationType" json:"notificationType" binding:"required"`
	EventType        string    `form:"eventType" json:"eventType" binding:"required"`
	CfgInfo          CfgInfo   `form:"cfgInfo" json:"cfgInfo"`
}

type CreateNotificationCfgForm struct {
	PageForm
	NotificationType string      `form:"notificationType" json:"notificationType" binding:"required"`
	EventType        string      `form:"eventType" json:"eventType" binding:"required"`
	UserIds          []models.Id `form:"userIds" json:"userIds"`
	CfgInfo          CfgInfo     `form:"cfgInfo" json:"cfgInfo"`
}

type DeleteNotificationCfgForm struct {
	PageForm
	Id models.Id `form:"id" json:"id" binding:"required"`
}
