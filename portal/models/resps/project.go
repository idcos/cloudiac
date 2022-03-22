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
