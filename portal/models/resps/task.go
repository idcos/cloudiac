// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package resps

import (
	"cloudiac/portal/models"
)

type TaskDetailResp struct {
	//desensitize.Task
	models.Task
	Creator string `json:"creator" example:"超级管理员"`
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

type TaskStepDetail struct {
	Id      models.Id    `json:"id"`
	Index   int          `json:"index"`
	Name    string       `json:"name"`
	TaskId  models.Id    `json:"taskId"`
	Status  string       `json:"status"`
	Message string       `json:"message"`
	StartAt *models.Time `json:"startAt"`
	EndAt   *models.Time `json:"endAt"`
	Type    string       `json:"type"`
}
