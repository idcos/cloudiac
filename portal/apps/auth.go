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
	"strings"
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
			if _, _, err := services.LdapAuthLogin(email, password); err != nil {
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
			username, dn, ldapErr := services.LdapAuthLogin(form.Email, form.Password)
			if ldapErr != nil {
				// 找不到账号时也返回 InvalidPassword 错误，避免暴露系统中己有用户账号
				c.Logger().Warnf("ldap auth login: %v", ldapErr)
				return nil, e.New(e.InvalidPassword, http.StatusBadRequest)
			}
			// 登录成功, 在用户表中添加该用户
			if err = createLdapUserAndRole(c, username, form.Email, dn); err != nil {
				c.Logger().Warnf("create user error: %v", err)
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	if err := validPassword(c, user, form.Email, form.Password); err != nil {
		return nil, e.New(e.InvalidPassword, http.StatusInternalServerError)
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

func createLdapUserAndRole(c *ctx.ServiceContext, username, email, dn string) e.Error {
	tx := c.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 登录成功, 在用户表中添加该用户
	user, err := services.CreateUser(tx, models.User{
		Name:  username,
		Email: email,
	})
	if err != nil {
		c.Logger().Warnf("create user error: %v", err)
		_ = tx.Rollback()
		return e.New(e.InternalError, http.StatusInternalServerError)
	}

	// 获取ldap用户的OU信息
	userOU := strings.TrimPrefix(dn, fmt.Sprintf("uid=%s,", username))

	// 根据OU获取组织权限
	ldapUserOrgOUs, err := services.GetLdapOUOrgByDN(tx, userOU)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// 更新用户组织权限
	err = services.RefreshUserOrgRoles(tx, user.Id, ldapUserOrgOUs)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// 根据OU获取项目权限
	ldapUserProjectOUs, err := services.GetLdapOUProjectByDN(tx, userOU)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// 更新用户项目权限
	err = services.RefreshUserProjectRoles(tx, user.Id, ldapUserProjectOUs)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("createLdapUserAndRole commit err: %s", err)
	}

	return nil
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
