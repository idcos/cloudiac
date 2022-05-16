// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/configs"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"fmt"
	"net/http"
	"time"
)

func validPassword(c *ctx.ServiceContext, user *models.User, email, password string) e.Error {
	valid, err := utils.CheckPassword(password, user.Password)
	if err != nil {
		c.Logger().Warnf("check password error: %v", err)
		return e.New(e.InternalError, http.StatusInternalServerError)
	}
	if !valid {
		// 如果本地账号验证未通过，进行ldap登陆验证
		if configs.Get().Ldap.LdapServer != "" {
			if _, err := services.LdapAuthLogin(email, password); err != nil {
				c.Logger().Warnf("login error: %v", err)
				return e.New(e.InvalidPassword, http.StatusUnauthorized)
			}
		}
	}
	return nil
}

// Login 用户登陆
func Login(c *ctx.ServiceContext, form *forms.LoginForm) (resp interface{}, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("user login: %s", form.Email))
	user, err := services.GetUserByEmail(c.DB(), form.Email)
	if err != nil {
		if err.Code() == e.UserNotExists && configs.Get().Ldap.LdapServer != "" {
			// 当错误为用户邮箱不存在的时候，尝试使用ldap 进行登录
			username, ldapErr := services.LdapAuthLogin(form.Email, form.Password)
			if ldapErr != nil {
				// 找不到账号时也返回 InvalidPassword 错误，避免暴露系统中己有用户账号
				c.Logger().Warnf("ldap auth login: %v", ldapErr)
				return nil, e.New(e.InvalidPassword, http.StatusBadRequest)
			}
			// 登录成功, 在用户表中添加该用户
			if user, err = services.CreateUser(c.DB(), models.User{
				Name:  username,
				Email: form.Email,
			}); err != nil {
				c.Logger().Warnf("create user error: %v", err)
				return nil, e.New(e.InternalError, http.StatusInternalServerError)
			}
		} else {
			return nil, err
		}
	}
	if er1 := validPassword(c, user, form.Email, form.Password); er1 != nil {
		return nil, err
	}
	token, er := services.GenerateToken(user.Id, user.Name, user.IsAdmin, 1*24*time.Hour)
	if er != nil {
		c.Logger().Errorf("name [%s] generateToken error: %v", user.Email, er)
		return nil, e.New(e.InvalidPassword, http.StatusBadRequest)
	}
	data := resps.LoginResp{
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

	data := resps.SsoResp{
		Token: token,
	}

	return data, err
}

// VerifySsoToken 验证 SSO token
func VerifySsoToken(c *ctx.ServiceContext, form *forms.VerifySsoTokenForm) (resp *resps.VerifySsoTokenResp, err e.Error) {
	user, err := services.VerifySsoToken(c.DB(), form.Token)
	if err != nil {
		return nil, err
	}

	return &resps.VerifySsoTokenResp{
		UserId: user.Id,
		Email:  user.Email,
	}, nil
}
