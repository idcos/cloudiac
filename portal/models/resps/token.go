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
