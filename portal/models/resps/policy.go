// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package resps

import (
	"cloudiac/portal/models"
	"cloudiac/portal/models/desensitize"
	"fmt"
)

type ScanResultPageResp struct {
	PolicyStatus string                `json:"policyStatus"` // 扫描状态
	Task         *desensitize.ScanTask `json:"task"`         // 扫描任务
	Total        int64                 `json:"total"`        // 总数
	PageSize     int                   `json:"pageSize"`     // 分页数量
	List         []*PolicyResultGroup  `json:"groups"`       // 策略组
}

type PolicyResultGroup struct {
	Id      models.Id      `json:"id"`
	Name    string         `json:"name"`
	Summary Summary        `json:"summary"`
	List    []PolicyResult `json:"list"` // 策略扫描结果
}

type PolicyResult struct {
	models.PolicyResult
	PolicyName      string `json:"policyName" example:"VPC 安全组规则"`  // 策略名称
	PolicySuppress  bool   `json:"policySuppress"`                  //是否屏蔽
	PolicyGroupName string `json:"policyGroupName" example:"安全策略组"` // 策略组名称
	FixSuggestion   string `json:"fixSuggestion" example:"建议您创建一个专有网络..."`
	Rego            string `json:"rego" example:""` //rego 代码文件内容
}

type Summary struct {
	Passed     int `json:"passed"`
	Violated   int `json:"violated"`
	Suppressed int `json:"suppressed"`
	Failed     int `json:"failed"`
}

type PolicyResp struct {
	models.Policy
	GroupName string `json:"groupName"`
	Creator   string `json:"creator"`
	Summary
}

type RespPolicyTpl struct {
	models.Template

	PolicyStatus string `json:"policyStatus"` // 策略检查状态, enum('passed','violated','pending','failed')

	PolicyGroups []NewPolicyGroup `json:"policyGroups" gorm:"-"`
	OrgName      string           `json:"orgName" form:"orgName" `
	Summary

	// 以下字段不返回
	Status string `json:"status,omitempty" gorm:"-" swaggerignore:"true"` // 模板状态(enabled, disable)
}

type PolicyErrorResp struct {
	models.PolicyResult
	TargetId     models.Id `json:"targetId"`
	EnvName      string    `json:"envName"`
	TemplateName string    `json:"templateName"`
}

func (PolicyErrorResp) TableName() string {
	return models.PolicyResult{}.TableName()
}

type PolicyScanReportResp struct {
	Total            PieChar         `json:"total"`            // 检测结果比例
	TaskScanCount    Polyline        `json:"scanCount"`        // 检测源执行次数
	PolicyScanCount  Polyline        `json:"policyScanCount"`  // 策略运行趋势
	PolicyPassedRate PolylinePercent `json:"policyPassedRate"` // 检测通过率趋势
}

type PieChar []PieSector

type PieSector struct {
	Name  string `json:"name" example:"08-20"`
	Value int    `json:"value" example:"10"`
}

type Polyline struct {
	Column []string `json:"column,omitempty" example:"08-20,08-21"`
	Value  []int    `json:"value,omitempty" example:"101,103"`
}

type PolylinePercent struct {
	Column []string  `json:"column,omitempty" example:"08-20,08-21"`
	Value  []Percent `json:"value,omitempty" example:"0.951,0.962"`
	Passed []int     `json:"-"`
	Total  []int     `json:"-"`
}

type Percent float64 // 百分比，保留1位百分比小数，0.951 = 95.1%

func (n Percent) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.3f", n)), nil
}

type PolicySummaryResp struct {
	ActivePolicy struct {
		Total   int     `json:"total"`   // 最近 15 天产生扫描记录的策略数量
		Last    int     `json:"last"`    // 16～30 天产生扫描记录的策略数量
		Changes float64 `json:"changes"` // 相较上次变化
		Summary PieChar `json:"summary"` // 策略状态包含
	} `json:"activePolicy"`

	UnresolvedPolicy struct {
		Total   int     `json:"total"`   // 最近 15 天产生扫描记录的策略数量
		Last    int     `json:"last"`    // 16～30 天产生扫描记录的策略数量
		Changes float64 `json:"changes"` // 相较上次变化： (total - last) / last
		Summary PieChar `json:"summary"` // 策略严重级别统计
	} `json:"unresolvedPolicy"`

	PolicyViolated      PieChar `json:"policyViolated"`      // 策略不通过
	PolicyGroupViolated PieChar `json:"policyGroupViolated"` // 策略组不通过
}

type ParseResp struct {
	Template *TfParse `json:"template"`
}

type PolicyTestResp struct {
	Data         interface{} `json:"data" swaggertype:"string" example:"{\n\"accurics\":{\n\"instanceWithNoVpc\":[\n{\n\"Id\":\"alicloud_instance.instance\"\n}\n]\n}\n}"` // 脚本测试输出，json文本
	Error        string      `json:"error" example:"1 error occurred: policy.rego:4: rego_parse_error: refs cannot be used for rule\n"`                                    // 脚本执行错误内容
	PolicyStatus string      `json:"policyStatus"`
}
