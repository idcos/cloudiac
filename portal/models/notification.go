// Copyright 2021 CloudJ Company Limited. All rights reserved.

package models

import (
	"cloudiac/portal/libs/db"
)

type NotificationCfg struct {
	BaseModel

	OrgId            Id     `json:"orgId" gorm:"size:32;not null;comment:组织ID"`
	NotificationType string `json:"notificationType" gorm:"type:enum('email','webhook');default:'email';comment:通知类型"`
	EventType        string `json:"eventType" gorm:"type:enum('all','failure');default:'failure';comment:事件类型"`
	UserId           Id     `json:"userId" gorm:"size:32;comment:用户ID"`
	CfgInfo          JSON   `json:"cfgInfo" gorm:"type:json;null;comment:通知配置"`
}

func (NotificationCfg) TableName() string {
	return "iac_org_notification_cfg"
}

func (o NotificationCfg) Migrate(sess *db.Session) (err error) {
	return nil
}
