package resps

import (
	"cloudiac/portal/models"
)

type RespPolicyEnv struct {
	models.Env

	PolicyStatus string `json:"policyStatus"` // 策略检查状态, enum('passed','violated','pending','failed')

	PolicyGroups []NewPolicyGroup `json:"policyGroups" gorm:"-"`
	Summary
	OrgName      string `json:"orgName" form:"orgName" `
	ProjectName  string `json:"projectName" form:"projectName" `
	TemplateName string `json:"templateName"`

	// 以下字段不返回
	Status     string `json:"status,omitempty" gorm:"-" swaggerignore:"true"`     // 环境状态
	TaskStatus string `json:"taskStatus,omitempty" gorm:"-" swaggerignore:"true"` // 环境部署任务状态
}

type RespEnvOfPolicy struct {
	models.Policy
	GroupName string `json:"groupName"`
	GroupId   string `json:"groupId"`
	EnvName   string `json:"envName"`
}

type ValidPolicyResp struct {
	ValidPolicies      []models.Policy `json:"validPolicies"`
	SuppressedPolicies []models.Policy `json:"suppressedPolicies"`
}
