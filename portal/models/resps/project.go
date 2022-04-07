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

type EnvDetailStatResp struct {
	Id    models.Id `json:"id"`
	Name  string    `json:"name"`
	Count int       `json:"count"`
}

type ProjectEnvStatResp struct {
	Status string              `json:"status"`
	Count  int                 `json:"count"`
	Envs   []EnvDetailStatResp `json:"envs"`
}

type EnvResStatResp struct {
	ResType string              `json:"resType"`
	Count   int                 `json:"count"`
	Envs    []EnvDetailStatResp `json:"envs"`
}

type EnvDetailStatWithUpResp struct {
	Id    models.Id `json:"id"`
	Name  string    `json:"name"`
	Count int       `json:"count"`
	Up    int       `json:"up"`
}

type ResTypeEnvetailStatWithUpResp struct {
	ResType string                    `json:"resType"`
	Count   int                       `json:"count"`
	Up      int                       `json:"up"`
	Envs    []EnvDetailStatWithUpResp `json:"envs"`
}

type ProjectEnvResStatResp struct {
	Date     string                          `json:"date"`
	ResTypes []ResTypeEnvetailStatWithUpResp `json:"ResTypes"`
}

type ProjectResGrowTrendResp struct {
	Date     string                          `json:"date"`
	Count    int                             `json:"count"`
	Up       int                             `json:"up"`
	ResTypes []ResTypeEnvetailStatWithUpResp `json:"ResTypes"`
}

type ProjectStatResp struct {
	EnvStat      []ProjectEnvStatResp        `json:"envStat"`
	ResStat      []EnvResStatResp            `json:"resStat"`
	EnvResStat   []ProjectEnvResStatResp     `json:"envResStat"`
	ResGrowTrend [][]ProjectResGrowTrendResp `json:"resGrowTrend"`
}
