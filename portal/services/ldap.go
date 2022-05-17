// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/configs"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/models/resps"
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

func connectLdap() (*ldap.Conn, e.Error) {
	conf := configs.Get()
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", conf.Ldap.LdapServer, conf.Ldap.LdapServerPort))
	if err != nil {
		return nil, e.New(e.LdapConnectFailed, err)
	}
	// 配置ldap 管理员dn信息，例如cn=Manager,dc=idcos,dc=com
	err = conn.Bind(conf.Ldap.AdminDn, conf.Ldap.AdminPassword)
	if err != nil {
		return nil, e.New(e.ValidateError, err)
	}

	return conn, nil
}

func closeLdap(conn *ldap.Conn) {
	if conn != nil {
		conn.Close()
	}
}

func SearchLdapOUs() ([]resps.LdapOUResp, e.Error) {
	conn, er := connectLdap()
	if er != nil {
		return nil, e.New(e.LdapConnectFailed, er)
	}
	defer closeLdap(conn)

	conf := configs.Get()
	// SearchFilter 需要内填入搜索条件，单个用括号包裹，例如 (objectClass=person)(!(userAccountControl=514))
	seachFilter := "(objectClass=organizationalUnit)"
	searchRequest := ldap.NewSearchRequest(
		conf.Ldap.OUSearchBase,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		seachFilter,
		// 这里是查询返回的属性,以数组形式提供.如果为空则会返回所有的属性
		[]string{"ou"},
		nil,
	)
	searchResults, err := conn.Search(searchRequest)
	if err != nil {
		return nil, e.New(e.ValidateError, err)
	}

	var results = make([]resps.LdapOUResp, 0)
	for _, sr := range searchResults.Entries {
		results = append(results, resps.LdapOUResp{
			DN: sr.DN,
			OU: sr.GetAttributeValue("ou"),
		})
	}

	return results, nil

}
