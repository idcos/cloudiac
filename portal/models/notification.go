// Copyright 2021 CloudJ Company Limited. All rights reserved.

package models

import (
	"cloudiac/portal/libs/db"
	"github.com/lib/pq"
)

const (
	NotificationTypeEmail      = "email"
	NotificationTypeWebhook    = "webhook"
	NotificationTypeWeChat     = "wechat"
	NotificationTypeSlack      = "slack"
	NotificationTypeDingTalk   = "dingtalk"
	NotificationEventFailed    = "failed"
	NotificationEventComplete  = "complete"
	NotificationEventRunning   = "running"
	NotificationEventApproving = "approving"
)

// 通知类型 email, webhook, 钉钉， 企业微信，slack
// 事件 running(发起)、approving(审批)、complete(成功)、failed(失败)

type Notification struct {
	BaseModel

	OrgId            Id     `json:"orgId" gorm:"size:32;not null;comment:组织ID"`
	ProjectId        Id     `json:"projectId" form:"projectId"  gorm:"size:32;not null;comment:项目ID"`
	Name             string `json:"name" form:"name" `
	NotificationType string `json:"notificationType" gorm:"type:enum('email', 'webhook', 'wechat', 'slack','dingtalk');default:'email';comment:通知类型"`
	//EventType        string         `json:"eventType" gorm:"type:enum('all', 'failed', 'complete', 'approving', 'running');default:'failure';comment:事件类型"`
	EventFailed    bool           `json:"eventFailed" gorm:"default:false"`
	EventComplete  bool           `json:"eventComplete" gorm:"default:false"`
	EventApproving bool           `json:"eventApproving" gorm:"default:false"`
	EventRunning   bool           `json:"eventRunning" gorm:"default:false"`
	Secret         string         `json:"secret" form:"secret" gorm:"comment:dingtalk加签秘钥"`
	Url            string         `json:"url" form:"url" gorm:"comment:回调url"`
	UserIds        pq.StringArray `json:"userIds" gorm:"size:32;comment:用户ID"  swaggertype:"array,string"`
}

func (Notification) TableName() string {
	return "iac_notification"
}

func (o Notification) Migrate(sess *db.Session) (err error) {
	return nil
}
