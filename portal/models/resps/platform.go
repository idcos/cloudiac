// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package resps

import "cloudiac/portal/models"

type BaseDataCount struct {
	Total  int64 `json:"total"`
	Active int64 `json:"active"`
}

type PfBasedataResp struct {
	OrgCount     BaseDataCount `json:"orgCount"`
	ProjectCount BaseDataCount `json:"projectCount"`
	EnvCount     BaseDataCount `json:"envCount"`
	StackCount   BaseDataCount `json:"stackCount"`
	UserCount    BaseDataCount `json:"userCount"`
}

type PfProEnvStatResp struct {
	Provider string `json:"provider"`
	Count    int64  `json:"count"`
}

type PfProResStatResp struct {
	Provider string `json:"provider"`
	Count    int64  `json:"count"`
}

type PfResTypeStatResp struct {
	ResType string `json:"resType"`
	Count   int64  `json:"count"`
}

type PfResWeekChangeResp struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type PfResTypeOrgsStatResp struct {
	ResType string  `json:"resType"`
	List    []int64 `json:"list"`
}

type PfActiveResStatResp struct {
	OrgList      []string                `json:"orgList"`
	ResTypesStat []PfResTypeOrgsStatResp `json:"resTypesStat"`
}

type OperationLogResp struct {
	models.UserOperationLog
	OperatorName string `json:"operatorName"`
	ActionName   string `json:"actionName"`
	OrgName      string `json:"orgName"`
}

type PfTodayStatResp struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type PfTodayResTypeStatResp struct {
	ResType string `json:"resType"`
	Count   int64  `json:"count"`
}
