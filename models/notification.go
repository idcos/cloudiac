package models

import (
	"cloudiac/libs/db"
)

type NotificationCfg struct {
	BaseModel

	OrgId            int    `json:"orgId" gorm:"not null;comment:'组织ID'"`
	NotificationType string `json:"notificationType" gorm:"size:32;not null;comment:'通知类型'"`
	EventType        string `json:"eventType" gorm:"size:32;comment:'事件类型'"`
	CfgInfo          JSON   `json:"cfgInfo" gorm:"type:json;null;comment:'通知配置'"`
}

func (NotificationCfg) TableName() string {
	return "iac_org_notification_cfg"
}

func (o NotificationCfg) Migrate(sess *db.Session) (err error) {
	if err != nil {
		return err
	}

	return nil
}
