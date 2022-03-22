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
}

type InviteUsersBatchResp struct {
	Success int `json:"success"`
	Failed  int `json:"failed"`
}
