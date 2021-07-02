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

type CreateOrganizationForm struct {
	BaseForm
	Name        string `form:"name" json:"name" binding:"required,gte=2,lte=32"` // 组织名称
	Description string `form:"description" json:"description" binding:""`        // 组织描述
	RunnerId    string `form:"runnerId" json:"runnerId" binding:""`              // 组织默认部署通道
}

type OrganizationParam struct {
	BaseForm

	Id models.Id `uri:"id" json:"id" binding:""` // 组织ID
}

type UpdateOrganizationForm struct {
	BaseForm
	Name        string `form:"name" json:"name" binding:""`                      // 组织名称
	Description string `form:"description" json:"description" binding:"max=255"` // 组织描述
	RunnerId    string `form:"runnerId" json:"runnerId" binding:""`              // 组织默认部署通道
	Status      string `form:"status" json:"status" enums:"enable,disable"`      // 组织状态
}

type SearchOrganizationForm struct {
	PageForm

	Q      string `form:"q" json:"q" binding:""`                       // 组织名称，支持模糊查询
	Status string `form:"status" json:"status" enums:"enable,disable"` // 组织状态
}

type DeleteOrganizationForm struct {
	BaseForm
}

type DisableOrganizationForm struct {
	BaseForm

	Status string `form:"status" json:"status" binding:"required"` // 组织状态
}

type DetailOrganizationForm struct {
	BaseForm
	Id models.Id `uri:"id" json:"id" binding:"required"` // 组织ID
}
