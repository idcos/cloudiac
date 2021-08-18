package models

type Violation struct {
	RuleName     string `json:"rule_name"`
	Description  string `json:"description"`
	RuleId       string `json:"rule_id"`
	Severity     string `json:"severity"`
	Category     string `json:"category"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	File         string `json:"file"`
	Line         int    `json:"line"`
}

type PolicyResult struct {
	AutoUintIdModel

	OrgId     Id `json:"org_id" gorm:"size:32;comment:组织" example:"org-c3lcrjxczjdywmk0go90"`
	ProjectId Id `json:"project_id" gorm:"size:32;comment:项目ID" example:"p-c3lcrjxczjdywmk0go90"`
	TplId     Id `json:"tpl_id" gorm:"size:32;comment:云模板ID" example:"tpl-c3lcrjxczjdywmk0go90"`
	EnvId     Id `json:"env_id" gorm:"size:32;comment:环境ID" example:"env-c3lcrjxczjdywmk0go90"`

	PolicyId      Id `json:"policyId" gorm:"size:32;comment:策略ID" example:"po-c3lcrjxczjdywmk0go90"`
	PolicyGroupId Id `json:"policyGroupId" gorm:"size:32;comment:策略组ID" example:"pog-c3lcrjxczjdywmk0go90"`

	StartAt *Time  `json:"startAt" gorm:"type:datetime;comment:任务开始时间"`                                                               // 任务开始时间
	Status  string `json:"status" gorm:"type:enum('passed','violated','suppressed','pending','failed');default:'pending';comment:状态"` // 策略扫描状态

	Violation
}

func (PolicyResult) TableName() string {
	return "iac_policy_result"
}
