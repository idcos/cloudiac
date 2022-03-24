// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"fmt"
	"net/http"
	"time"
)

func validPassword(password, userPassword string) e.Error {
	valid, er := utils.CheckPassword(password, userPassword)
	if er != nil {
		return e.New(e.ValidateError, http.StatusInternalServerError, er)
	}
	if !valid {
		return e.New(e.InvalidPassword, http.StatusBadRequest)
	}
	return nil
}

// Login 用户登陆
func Login(c *ctx.ServiceContext, form *forms.LoginForm) (resp interface{}, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("user login: %s", form.Email))
	user, err := services.GetUserByEmail(c.DB(), form.Email)
	// ldap 账号只能使用ldap账号进行登录
	if err == nil {
		if user.IsLdap {
			if _, err = services.LdapAuthLogin(form.Email, form.Password); err != nil {
				return nil, err
			}
		} else {
			if err = validPassword(form.Password, user.Password); err != nil {
				return nil, err
			}
		}
	} else {
		if err.Code() == e.UserNotExists {
			// 使用ldap 进行登录
			username, ldapErr := services.LdapAuthLogin(form.Email, form.Password)
			if ldapErr != nil {
				// 找不到账号时也返回 InvalidPassword 错误，避免暴露系统中己有用户账号
				c.Logger().Error(ldapErr)
				return nil, e.New(e.InvalidPassword, http.StatusBadRequest)
			}
			// 登录成功, 标记账号为ldap用户，并且在用户表中添加该用户
			createUserform := &forms.CreateUserForm{
				Name:   username,
				Email:  form.Email,
				IsLdap: true,
			}
			ldapUser, err := CreateUser(c, createUserform)
			if err != nil {
				return nil, err
			}
			user = ldapUser.User
		} else {
			return nil, e.New(e.DBError, err)
		}
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
