// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package resps

import "cloudiac/portal/models"

type RespNotification struct {
	models.Notification
	EventType   models.Text `json:"-" form:"-" gorm:"event_type"`
	EventTypes  []string    `json:"eventType" form:"eventType" gorm:"-"`
	CreatorName string      `json:"creatorName" form:"creatorName" `
}

type RespDetailNotification struct {
	models.Notification
	EventType  string   `json:"-" `
	EventTypes []string `json:"eventType" gorm:"-"`
}
