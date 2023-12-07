// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

type PolicyResult struct {
	AutoUintIdModel

	OrgId     Id `json:"org_id" gorm:"not null;size:32;comment:组织" example:"org-c3lcrjxczjdywmk0go90"` // 组织ID
	ProjectId Id `json:"project_id" gorm:"size:32;comment:项目ID" example:"p-c3lcrjxczjdywmk0go90"`      // 项目ID
	TplId     Id `json:"tpl_id" gorm:"size:32;comment:云模板ID" example:"tpl-c3lcrjxczjdywmk0go90"`       // 云模板ID
	EnvId     Id `json:"env_id" gorm:"size:32;comment:环境ID" example:"env-c3lcrjxczjdywmk0go90"`        // 环境ID

	TaskId Id `json:"taskId" gorm:"not null;size:32;index;comment:任务ID" example:"t-c3lcrjxczjdywmk0go90"` // 任务ID

	PolicyId      Id `json:"policyId" gorm:"not null;size:32;comment:策略ID" example:"po-c3lcrjxczjdywmk0go90"`        // 策略ID
	PolicyGroupId Id `json:"policyGroupId" gorm:"not null;size:32;comment:策略组ID" example:"pog-c3lcrjxczjdywmk0go90"` // 策略组ID

	StartAt Time `json:"startAt" gorm:"type:datetime;index;comment:开始时间"` // 任务开始时间

	Status  string `json:"status" gorm:"default:'pending';comment:状态"` // 状态 type:enum('passed','violated','suppressed','pending','failed');
	Message Text   `json:"message" gorm:"type:text;comment:失败原因"`

	Violation
}

func (PolicyResult) TableName() string {
	return "iac_policy_result"
}

type TsResultJson struct {
	Results TsResult `json:"results"`
}

type TsResult struct {
	ScanErrors        []ScanError `json:"scan_errors,omitempty"`
	PassedRules       []Rule      `json:"passed_rules,omitempty"`
	Violations        []Violation `json:"violations"`
	SkippedViolations []Violation `json:"skipped_violations"`
	ScanSummary       ScanSummary `json:"scan_summary"`
}

type ScanError struct {
	IacType   string `json:"iac_type"`
	Directory string `json:"directory"`
	ErrMsg    string `json:"errMsg"`
}

type ScanSummary struct {
	FileFolder        string `json:"file/folder"`
	IacType           string `json:"iac_type"`
	ScannedAt         string `json:"scanned_at"`
	PoliciesValidated int    `json:"policies_validated"`
	ViolatedPolicies  int    `json:"violated_policies"`
	Low               int    `json:"low"`
	Medium            int    `json:"medium"`
	High              int    `json:"high"`
}

type Rule struct {
	RuleName    string `json:"rule_name"`   // 规则名称
	Description string `json:"description"` // 规则描述
	RuleId      string `json:"rule_id"`     // 规则ID（策略ID）
	Severity    string `json:"severity"`    // 严重程度
	Category    string `json:"category"`    // 分类（策略组名称）
}

type Violation struct {
	RuleName     string `json:"rule_name" gorm:"comment:策略名称"`                  // 规则名称
	Description  string `json:"description" gorm:"comment:策略描述"`                // 规则描述
	RuleId       string `json:"rule_id" gorm:"comment:规则ID(策略ID)"`              // 规则ID（策略ID）
	Severity     string `json:"severity" gorm:"comment:严重程度"`                   // 严重程度
	Category     string `json:"category" gorm:"comment:分类（策略组名称）"`              // 分类（策略组名称）
	Comment      string `json:"skip_comment,omitempty" gorm:"comment:跳过说明"`     // 注释
	ResourceName string `json:"resource_name" gorm:"comment:资源名称"`              // 资源名称
	ResourceType string `json:"resource_type" gorm:"comment:资源类型"`              // 资源类型
	ModuleName   string `json:"module_name,omitempty" gorm:"comment:模块名称"`      // 模块名称
	File         string `json:"file,omitempty" gorm:"comment:源码文件名"`            // 文件路径
	PlanRoot     string `json:"plan_root,omitempty" gorm:"comment:源码文件夹"`       // 文件夹路径
	Line         int    `json:"line,omitempty" gorm:"comment:错误资源源码行号"`         // 错误源文件行号
	Source       Text   `json:"source,omitempty" gorm:"type:text;comment:错误源码"` // 错误源码
}

type TsCount struct {
	Low    int `json:"low"`
	Medium int `json:"medium"`
	High   int `json:"high"`
	Total  int `json:"total"`
}

type TSResource struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	ModuleName string `json:"module_name"`
	Source     string `json:"source"`
	PlanRoot   string `json:"plan_root"`
	Line       int    `json:"line"`
	Type       string `json:"type"`

	Config map[string]interface{} `json:"config"`

	SkipRules   *bool  `json:"skip_rules"`
	MaxSeverity string `json:"max_severity"`
	MinSeverity string `json:"min_severity"`
}

type TSResources []TSResource

type TfParse map[string]TSResources
