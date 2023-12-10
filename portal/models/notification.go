// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/portal/libs/db"
)

const (
	NotificationTypeEmail    = "email"
	NotificationTypeWebhook  = "webhook"
	NotificationTypeWeChat   = "wechat"
	NotificationTypeSlack    = "slack"
	NotificationTypeDingTalk = "dingtalk"
)

// 通知类型 email, webhook, 钉钉， 企业微信，slack
// 事件 running(发起)、approving(审批)、complete(成功)、failed(失败)

type Notification struct {
	BaseModel

	OrgId     Id          `json:"orgId" gorm:"size:32;not null;comment:组织ID"`
	ProjectId Id          `json:"projectId" form:"projectId"  gorm:"size:32;not null;comment:项目ID"`
	Name      string      `json:"name" form:"name" `
	Type      string      `json:"notificationType" gorm:"default:'email';comment:通知类型"` // type:enum('email', 'webhook', 'wechat', 'slack','dingtalk');
	Secret    string      `json:"secret" form:"secret" gorm:"comment:dingtalk加签秘钥"`
	Url       string      `json:"url" form:"url" gorm:"comment:回调url"`
	UserIds   StringArray `json:"userIds"  gorm:"type:text;comment:用户ID"  swaggertype:"array,string"`
	Creator   Id          `json:"creator" form:"creator" `
}

func (Notification) TableName() string {
	return "iac_notification"
}

type NotificationEvent struct {
	AutoUintIdModel

	EventType      string `json:"eventType" form:"eventType"  gorm:"default:'task.running';comment:事件类型"` // type:enum('task.failed', 'task.complete', 'task.approving', 'task.running', 'task.crondrift');
	NotificationId Id     `json:"notificationId" form:"notificationId" gorm:"size:32;not null"`
}

func (NotificationEvent) TableName() string {
	return "iac_notification_event"
}

func (NotificationEvent) Migrate(tx *db.Session) error {
	if err := tx.ModifyModelColumn(&NotificationEvent{}, "event_type"); err != nil {
		return err
	}
	return nil
}
