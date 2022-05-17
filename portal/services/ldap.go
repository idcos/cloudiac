// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/configs"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
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

func SearchLdapOUs() (*resps.LdapOUResp, e.Error) {
	conn, er := connectLdap()
	if er != nil {
		return nil, e.New(e.LdapConnectFailed, er)
	}
	defer closeLdap(conn)

	conf := configs.Get()
	seachFilter := "(objectClass=organizationalUnit)"
	searchRequest := ldap.NewSearchRequest(
		conf.Ldap.OUSearchBase,
		ldap.ScopeBaseObject, ldap.DerefAlways, 0, 0, false,
		seachFilter,
		// 这里是查询返回的属性,以数组形式提供.如果为空则会返回所有的属性
		[]string{"ou"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, e.New(e.ValidateError, err)
	}

	if len(sr.Entries) == 0 {
		return nil, e.New(e.ObjectNotExists, err)
	}

	var root = &resps.LdapOUResp{
		DN: sr.Entries[0].DN,
		OU: sr.Entries[0].GetAttributeValue("ou"),
	}

	err = genOUTree(conn, root)
	if err != nil {
		return nil, e.New(e.ValidateError, err)
	}

	return root, nil
}

func genOUTree(conn *ldap.Conn, root *resps.LdapOUResp) error {
	seachFilter := "(objectClass=organizationalUnit)"
	searchRequest := ldap.NewSearchRequest(
		root.DN,
		ldap.ScopeSingleLevel, ldap.DerefAlways, 0, 0, false,
		seachFilter,
		// 这里是查询返回的属性,以数组形式提供.如果为空则会返回所有的属性
		[]string{"ou"},
		nil,
	)

	searchResults, err := conn.Search(searchRequest)
	if err != nil {
		return err
	}

	children := make([]resps.LdapOUResp, 0)
	for _, sr := range searchResults.Entries {
		child := resps.LdapOUResp{
			DN: sr.DN,
			OU: sr.GetAttributeValue("ou"),
		}
		err = genOUTree(conn, &child)
		if err != nil {
			return err
		}

		children = append(children, child)
	}

	root.Children = children

	return nil
}

func SearchLdapUsers(q string, count int) ([]resps.LdapUserResp, e.Error) {
	conn, er := connectLdap()
	if er != nil {
		return nil, e.New(e.LdapConnectFailed, er)
	}
	defer closeLdap(conn)

	conf := configs.Get()
	// SearchFilter 需要内填入搜索条件，单个用括号包裹，例如 (objectClass=person)(!(userAccountControl=514))
	seachFilter := fmt.Sprintf("(&%s(%s=%s))", conf.Ldap.SearchFilter, conf.Ldap.AccountAttribute, "*")
	searchRequest := ldap.NewSearchRequest(
		conf.Ldap.SearchBase,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		seachFilter,
		// 这里是查询返回的属性,以数组形式提供.如果为空则会返回所有的属性
		[]string{},
		nil,
	)
	searchResults, err := conn.Search(searchRequest)
	if err != nil {
		return nil, e.New(e.ValidateError, err)
	}

	var results = make([]resps.LdapUserResp, 0)
	for _, sr := range searchResults.Entries {
		results = append(results, resps.LdapUserResp{
			DN:    sr.DN,
			Email: sr.GetAttributeValue(conf.Ldap.EmailAttribute),
			Uid:   sr.GetAttributeValue(conf.Ldap.AccountAttribute),
		})
	}

	return results, nil
}

func CreateOUOrg(sess *db.Session, m models.LdapOUOrg) (*resps.AuthLdapOUResp, e.Error) {
	err := sess.Insert(&m)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return &resps.AuthLdapOUResp{
		Id: m.Id.String(),
	}, nil
}
