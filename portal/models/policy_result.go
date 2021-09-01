package models

type PolicyResult struct {
	AutoUintIdModel

	OrgId     Id `json:"org_id" gorm:"not null;size:32;comment:组织" example:"org-c3lcrjxczjdywmk0go90"`
	ProjectId Id `json:"project_id" gorm:"size:32;comment:项目ID" example:"p-c3lcrjxczjdywmk0go90"`
	TplId     Id `json:"tpl_id" gorm:"size:32;comment:云模板ID" example:"tpl-c3lcrjxczjdywmk0go90"`
	EnvId     Id `json:"env_id" gorm:"size:32;comment:环境ID" example:"env-c3lcrjxczjdywmk0go90"`
	TaskId    Id `json:"taskId" gorm:"not null;size:32;comment:任务ID" example:"t-c3lcrjxczjdywmk0go90"`

	PolicyId      Id `json:"policyId" gorm:"not null;size:32;comment:策略ID" example:"po-c3lcrjxczjdywmk0go90"`
	PolicyGroupId Id `json:"policyGroupId" gorm:"not null;size:32;comment:策略组ID" example:"pog-c3lcrjxczjdywmk0go90"`

	StartAt Time   `json:"startAt" gorm:"type:datetime;comment:开始时间"`                                                                 // 任务开始时间
	Status  string `json:"status" gorm:"type:enum('passed','violated','suppressed','pending','failed');default:'pending';comment:状态"` // 策略扫描状态

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
	RuleName    string `json:"rule_name"`
	Description string `json:"description"`
	RuleId      string `json:"rule_id"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
}

type Violation struct {
	RuleName     string `json:"rule_name"`
	Description  string `json:"description"`
	RuleId       string `json:"rule_id"`
	Severity     string `json:"severity"`
	Category     string `json:"category"`
	Comment      string `json:"skip_comment,omitempty"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	ModuleName   string `json:"module_name,omitempty"`
	File         string `json:"file,omitempty"`
	PlanRoot     string `json:"plan_root,omitempty"`
	Line         int    `json:"line,omitempty"`
	Source       string `json:"source,omitempty"`
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
