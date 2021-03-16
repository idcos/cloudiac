package services

import (
	"cloudiac/configs"
	"cloudiac/libs/ctx"
	"cloudiac/utils/rabbitmq"
)

type OperateLog struct {
	Module              string                `json:"module"`
	EventType           string                `json:"eventType"`
	ResourceType        string                `json:"resourceType"`
	ResourceName        string                `json:"resourceName"`
	TenantId            uint                  `json:"tenantId"`
	SourceIpAddress     string                `json:"sourceIpAddress"`
	UserAgent           string                `json:"userAgent"`
	UserIdentity        UserIdentity          `json:"userIdentity"`
	ReferencedResources []ReferencedResources `json:"referencedResources"`
}

type UserIdentity struct {
	Type     string `json:"type"`
	Username string `json:"username"`
}

type ReferencedResources struct {
	ResourceType string `json:"resourceType"`
	ResourceName string `json:"resourceName"`
}

func SendOperateLog(eventType string, resourceType string, resourceName string, c *ctx.ServiceCtx) {
	if rabbitmq.MQ == nil {
		return
	}

	module := "monitor"
	data := &OperateLog{
		Module:       module,
		EventType:    eventType,
		ResourceType: resourceType,
		ResourceName: resourceName,
		UserIdentity: UserIdentity{
			Type:     "user",
			Username: c.Username,
		},
		TenantId:        c.TenantId,
		SourceIpAddress: c.UserIpAddr,
		UserAgent:       c.UserAgent,
	}
	go rabbitmq.MQ.Send(configs.Get().Rmq.Queue, data)
}
