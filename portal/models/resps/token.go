// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package resps

import "cloudiac/portal/models"

type LoginResp struct {
	//UserInfo *models.User
	Token string `json:"token" example:"eyJhbGciO..."` // 登陆令牌
}

type SsoResp struct {
	Token string `json:"token" example:"eyJhbGciO..."` // SSO令牌
}

type VerifySsoTokenResp struct {
	UserId models.Id `json:"userId"`
	Email  string    `json:"email"`
}

type TokenResp struct {
	models.TimedModel

	Name        string       `json:"name"`
	Type        string       `json:"type"`
	OrgId       models.Id    `json:"orgId"`
	Role        string       `json:"role"`
	Status      string       `json:"status"`
	ExpiredAt   *models.Time `json:"expiredAt"`
	Description string       `json:"description"`
	CreatorId   models.Id    `json:"creatorId" example:"u-c3ek0co6n88ldvq1n6ag"`

	// 触发器需要的字段
	EnvId  models.Id `json:"envId"`
	Action string    `json:"action"`
	Key    string    `json:"key"`
}
