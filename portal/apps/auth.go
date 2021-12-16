// Copyright 2021 CloudJ Company Limited. All rights reserved.

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

	"github.com/dgrijalva/jwt-go"
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

type SsoTokenClaims struct {
	jwt.StandardClaims

	UserId models.Id `json:"userId"`
}

func GenerateSsoToken(c *ctx.ServiceContext) (resp *SsoTokenClaims, err e.Error) {
	// TODO
	return nil, nil
}

type VerifySsoTokenForm struct {
	forms.BaseForm

	Token string `json:"token" form:"token" binding:"required"`
}

type VerifySsoTokenResp struct {
	UserId   string
	Username string
}

func VerifySsoToken(c *ctx.ServiceContext, form *VerifySsoTokenForm) (resp *VerifySsoTokenResp, err e.Error) {
	// TODO
	return nil, nil
}
