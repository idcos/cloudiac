// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package resps

import "cloudiac/portal/models"

type ProjectResp struct {
	models.Project
	Creator string `json:"creator" form:"creator" `
}

type DetailProjectResp struct {
	models.Project
	UserAuthorization []models.UserProject `json:"userAuthorization" form:"userAuthorization" ` //用户认证信息
	ProjectStatistics
}

type ProjectStatistics struct {
	TplCount    int64 `json:"tplCount" form:"tplCount" `
	EnvActive   int64 `json:"envActive" form:"envActive" `
	EnvFailed   int64 `json:"envFailed" form:"envFailed" `
	EnvInactive int64 `json:"envInactive" form:"envInactive" `
}

type EnvResStatResp struct {
	EnvId   string `json:"envId"`
	EnvName string `json:"envName"`
	ResType string `json:"resType"`
	Date    string `json:"date"`
	Count   int    `json:"count"`
}

type EnvResSummaryResp struct {
	ResType string `json:"resType"`
	Count   int    `json:"count"`
	Up      int    `json:"up"` // 增长数量
}

type ProjectStatResp struct {
	EnvStat       []EnvStatResp       `json:"envStat"`
	ResStat       []ResStatResp       `json:"resStat"`
	EnvResStat    []EnvResStatResp    `json:"envResStat"`
	ResGrowTrend  []ResGrowTrendResp  `json:"resGrowTrend"`
	EnvResSummary []EnvResSummaryResp `json:"envResSummary"`
}
