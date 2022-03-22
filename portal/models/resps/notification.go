package resps

import "cloudiac/portal/models"

type RespNotification struct {
	models.Notification
	EventType   string   `json:"-" form:"-" gorm:"event_type"`
	EventTypes  []string `json:"eventType" form:"eventType" gorm:"-"`
	CreatorName string   `json:"creatorName" form:"creatorName" `
}

type RespDetailNotification struct {
	models.Notification
	EventType  string   `json:"-" `
	EventTypes []string `json:"eventType" gorm:"-"`
}
