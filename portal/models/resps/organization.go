// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package resps

import "cloudiac/portal/models"

type OrgDetailResp struct {
	models.Organization
	Creator string `json:"creator" example:"研发部负责人"` // 创建人名称
}

type OrganizationDetailResp struct {
	models.Organization
	Creator string `json:"creator" example:"超级管理员"`
}

type OrgResourcesResp struct {
	ProjectName  string    `json:"projectName"`
	EnvName      string    `json:"envName"`
	ResourceName string    `json:"resourceName"`
	Provider     string    `json:"provider"`
	Type         string    `json:"type"`
	Module       string    `json:"module"`
	EnvId        models.Id `json:"envId"`
	ProjectId    models.Id `json:"projectId"`
	ResourceId   models.Id `json:"resourceId"`
	Attrs        string    `json:"attrs"`
}

type InviteUsersBatchResp struct {
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

type EnvResp struct {
	EnvName string    `json:"envName"`
	EnvId   models.Id `json:"envId"`
}
type OrgEnvAndProviderResp struct {
	Envs      []EnvResp `json:"envs"`
	Providers []string  `json:""`
}

type EnvStatResp struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type ResStatResp struct {
	ResType string `json:"resType"`
	Count   int    `json:"count"`
}

type ProjectStatResp struct {
	ProjectId   string `json:"projectId"`
	ProjectName string `json:"projectName"`
	ResType     string `json:"resType"`
	Date        string `json:"date"`
	Count       int    `json:"count"`
}

type ResGrowTrendResp struct {
	ProjectId   string `json:"projectId"`
	ProjectName string `json:"projectName"`
	Date        string `json:"date"`
	Count       int    `json:"count"`
}

type OrgProjectsStatResp struct {
	EnvStat      []EnvStatResp      `json:"envStat"`
	ResStat      []ResStatResp      `json:"resStat"`
	ProjectStat  []ProjectStatResp  `json:"projectStat"`
	ResGrowTrend []ResGrowTrendResp `json:"resGrowTrend"`
}
