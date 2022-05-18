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
	"gorm.io/gorm"
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
	// 判断ou是否存在
	var ouOrg models.LdapOUOrg
	err := sess.Model(&models.LdapOUOrg{}).Where(`org_id = ?`, m.OrgId).Where(`dn = ?`, m.DN).First(&ouOrg)

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, e.New(e.DBError, err)
	}

	if err == gorm.ErrRecordNotFound {
		err = sess.Insert(&m)
	} else {
		m.Id = ouOrg.Id
		_, err = sess.Model(&ouOrg).Update(models.LdapOUOrg{Role: m.Role})
	}

	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return &resps.AuthLdapOUResp{
		Id: m.Id.String(),
	}, nil
}

func CreateLdapUserOrg(sess *db.Session, orgId models.Id, m models.User, role string) (*resps.AuthLdapUserResp, e.Error) {
	var err error
	// 判断 user 是否存在
	var user models.User
	err = sess.Model(&models.User{}).Where(`email = ?`, m.Email).First(&user)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, e.New(e.DBError, err)
	}

	tx := sess.Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 用户不存在
	userId := user.Id
	if err == gorm.ErrRecordNotFound {
		m.Id = models.NewId("u")
		err = tx.Insert(&m)
		if err != nil {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, err)
		}

		userId = m.Id
	}

	// 用户授权不存在
	var userOrg models.UserOrg
	err = sess.Model(&models.UserOrg{}).Where("user_id = ? and org_id = ?", userId, orgId).First(&userOrg)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, e.New(e.DBError, err)
	}

	if err == gorm.ErrRecordNotFound {
		err = sess.Insert(&models.UserOrg{
			UserId: userId,
			OrgId:  orgId,
			Role:   role,
		})
	} else {
		_, err = sess.Model(&userOrg).Update(models.UserOrg{OrgId: orgId, UserId: userId, Role: role})
	}

	if err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return &resps.AuthLdapUserResp{
		Id: string(userId),
	}, nil
}

func GetOrgLdapOUs(sess *db.Session, orgId models.Id) ([]resps.OrgLdapOUsResp, e.Error) {
	var results = make([]resps.OrgLdapOUsResp, 0)
	err := sess.Model(&models.LdapOUOrg{}).Select("dn", "ou").Find(&results)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return results, nil
}

func CreateOUProject(sess *db.Session, m models.LdapOUProject) (*resps.AuthLdapOUResp, e.Error) {
	// 判断ou是否存在
	var ouProject models.LdapOUProject
	err := sess.Model(&models.LdapOUProject{}).Where(`org_id = ?`, m.OrgId).Where(`dn = ?`, m.DN).Where(`project_id = ?`, m.ProjectId).First(&ouProject)

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, e.New(e.DBError, err)
	}

	if err == gorm.ErrRecordNotFound {
		err = sess.Insert(&m)
	} else {
		m.Id = ouProject.Id
		_, err = sess.Model(&ouProject).Update(models.LdapOUProject{Role: m.Role})
	}

	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return &resps.AuthLdapOUResp{
		Id: m.Id.String(),
	}, nil
}
