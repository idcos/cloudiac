// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package resps

import "cloudiac/portal/models"

type PolicyGroupResp struct {
	models.PolicyGroup
	PolicyCount uint     `json:"policyCount" example:"10"`
	RelCount    uint     `json:"relCount"`
	Labels      []string `json:"labels" gorm:"-"`
}

type PolicyGroupScanReportResp struct {
	PassedRate PolylinePercent `json:"passedRate"` // 检测通过率
}

type LastScanTaskResp struct {
	models.ScanTask
	TargetName  string `json:"targetName"`  // 检查目标
	TargetType  string `json:"targetType"`  // 目标类型：环境/模板
	OrgName     string `json:"orgName"`     // 组织名称
	ProjectName string `json:"projectName"` // 项目
	Creator     string `json:"creator"`     // 创建者
	Summary
}

type TemplateChecksResp struct {
	CheckResult string `json:"CheckResult"`
	Reason      string `json:"reason"`
}

type RegistryPGResp struct {
	VcsId     string `json:"vcsId"`
	RepoId    string `json:"repoId"`
	Namespace string `json:"namespace"`
	GroupId   string `json:"groupId"`
	GroupName string `json:"groupName"`
	Label     string `json:"label"`
}

type RegistryPGVerResp struct {
	Namespace string   `json:"namespace"`
	GroupName string   `json:"groupName"`
	GitTags   []string `json:"gitTags"`
}

type NewPolicyGroup struct {
	models.PolicyGroup
	OrgId     models.Id `json:"orgId"`
	ProjectId models.Id `json:"projectId" `
	TplId     models.Id `json:"tplId"`
	EnvId     models.Id `json:"envId"`
	Scope     string    `json:"scope"`
}
