// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"fmt"
	"net/http"
	"time"
)

// Login 用户登陆
func Login(c *ctx.ServiceContext, form *forms.LoginForm) (resp interface{}, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("user login: %s", form.Email))

	user, err := services.GetUserByEmail(c.DB(), form.Email)
	if err != nil {
		if err.Code() == e.UserNotExists {
			// 找不到账号时也返回 InvalidPassword 错误，避免暴露系统中己有用户账号
			return nil, e.New(e.InvalidPassword, http.StatusBadRequest)
		}
		return nil, e.New(e.DBError, err)
	}

	valid, er := utils.CheckPassword(form.Password, user.Password)
	if er != nil {
		return nil, e.New(e.ValidateError, http.StatusInternalServerError, er)
	}
	if !valid {
		return nil, e.New(e.InvalidPassword, http.StatusBadRequest)
	}

	token, er := services.GenerateToken(user.Id, user.Name, user.IsAdmin, 1*24*time.Hour)
	if er != nil {
		c.Logger().Errorf("name [%s] generateToken error: %v", user.Email, er)
		return nil, e.New(e.InvalidPassword, http.StatusBadRequest)
	}

	data := models.LoginResp{
		//UserInfo: user,
		Token: token,
	}

	return data, nil
}

// GenerateSsoToken 生成 SSO token
func GenerateSsoToken(c *ctx.ServiceContext) (resp interface{}, err e.Error) {

	token, er := services.GenerateSsoToken(c.UserId, 5*time.Minute)
	if er != nil {
		c.Logger().Errorf("userId [%s] generateToken error: %v", c.UserId, er)
		return nil, e.New(e.InternalError, er, http.StatusInternalServerError)
	}

	data := models.SsoResp{
		Token: token,
	}

	return data, err
}

// VerifySsoToken 验证 SSO token
func VerifySsoToken(c *ctx.ServiceContext, form *forms.VerifySsoTokenForm) (resp *models.VerifySsoTokenResp, err e.Error) {
	user, err := services.VerifySsoToken(c.DB(), form.Token)
	if err != nil {
		return nil, err
	}

	return &models.VerifySsoTokenResp{
		UserId: user.Id,
		Email:  user.Email,
	}, nil
}
