// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package resps

import "cloudiac/portal/models"

type NewVariable []VariableResp

func (v NewVariable) Len() int {
	return len(v)
}
func (v NewVariable) Less(i, j int) bool {
	return v[i].Name < v[j].Name
}
func (v NewVariable) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

type VariableResp struct {
	models.Variable
	Overwrites *models.Variable `json:"overwrites" form:"overwrites" ` //回滚参数，无需回滚是为空
}

type SearchVariableGroupRespTemp struct {
	models.VariableGroup
	Creator     string `json:"creator"`
	ProjectName string `json:"projectName,omitempty"`
}

//ProjectNames []string `json:"ProjectNames"`

type SearchVariableGroupResp struct {
	SearchVariableGroupRespTemp
	ProjectNames []string `json:"projectNames"`
}

type CreateVariableGroupResp struct {
	models.VariableGroup
	ProjectIds []models.Id `json:"projectIds"`
}

type DetailVariableGroupRespTemp struct {
	models.VariableGroup
	ProjectName string `json:"projectName,omitempty"`
	ProjectId   string `json:"projectId,omitempty"`
}

type DetailVariableGroupResp struct {
	DetailVariableGroupRespTemp
	ProjectNames []string `json:"projectNames"`
	ProjectIds   []string `json:"projectIds"`
}
