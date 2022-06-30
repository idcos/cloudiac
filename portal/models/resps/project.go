// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package resps

import "cloudiac/portal/models"

type ProjectResp struct {
	models.Project
	Creator           string               `json:"creator" form:"creator"`
	ActiveEnvironment int                  `json:"activeEnvironment"`
	ResStats          []ProjectResStatResp `json:"resStats" gorm:"-"`
}

type ProjectResStatResp struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
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

type ProjectStatResp struct {
	EnvStat      []EnvStatResp          `json:"envStat"`
	ResStat      []ResStatResp          `json:"resStat"`
	EnvResStat   []ProjOrEnvResStatResp `json:"envResStat"`
	ResGrowTrend []ResGrowTrendResp     `json:"resGrowTrend"`
}
