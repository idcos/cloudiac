// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/configs"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"log"
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

// 处理Ldap 登录逻辑
func LdapAuthLogin(username, password string) e.Error {
	conf := configs.Get()
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", conf.Ldap.LdapServer, conf.Ldap.LdapServerPort))
	if err != nil {
		return e.New(e.LdapConnectFailed, err)
	}
	defer conn.Close()
	// 配置ldap 管理员dn信息，例如cn=Manager,dc=idcos,dc=com
	err = conn.Bind(conf.Ldap.AdminDn, conf.Ldap.AdminPassword)
	if err != nil {
		return e.New(e.ValidateError, err)
	}

	searchRequest := ldap.NewSearchRequest(
		// 这里是 basedn,我们将从这个节点开始搜索
		"dc=idcos,dc=com",
		// 这里几个参数分别是 scope, derefAliases, sizeLimit, timeLimit,  typesOnly
		// 详情可以参考 RFC4511 中的定义,文末有链接
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		// 这里是 LDAP 查询的 Filter.这个例子例子,我们通过查询 uid=username 且 objectClass=organizationalPerson.
		// username 即我们需要认证的用户名
		fmt.Sprintf("(&(objectClass=organizationalPerson)(uid=%s))","okr"),
		// 这里是查询返回的属性,以数组形式提供.如果为空则会返回所有的属性
		[]string{"dn"},
		nil,
	)
	sr, err := conn.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("这里", sr)
	userdn := sr.Entries[0].DN
	fmt.Println("userdn", userdn)
	err = conn.Bind(userdn, "123456")
	if err != nil {
		fmt.Println("333", err)
		log.Fatal(err)
	}
	fmt.Println("成功")

}