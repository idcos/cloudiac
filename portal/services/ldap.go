// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/configs"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/resps"
	"errors"
	"fmt"

	"github.com/go-ldap/ldap/v3"
	"gorm.io/gorm"
)

func connectLdap() (*ldap.Conn, e.Error) {
	conf := configs.Get()

	// ldap 未配置
	if !conf.LdapEnabled() {
		return nil, e.New(e.LdapNotExisted, fmt.Errorf("ldap 未配置"))
	}

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
		if er.Code() == e.LdapNotExisted {
			return nil, nil
		}
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

func GetLdapUserByEmail(emails []string) ([]*models.User, e.Error) {
	conn, er := connectLdap()
	if er != nil {
		if er.Code() == e.LdapNotExisted {
			return nil, nil
		}
		return nil, e.New(e.LdapConnectFailed, er)
	}
	defer closeLdap(conn)

	conf := configs.Get()
	users := make([]*models.User, 0)

	for _, email := range emails {
		seachFilter := fmt.Sprintf("(&%s(%s=%s))", conf.Ldap.SearchFilter, conf.Ldap.EmailAttribute, email)
		searchRequest := ldap.NewSearchRequest(
			conf.Ldap.SearchBase,
			ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
			seachFilter,
			// 这里是查询返回的属性,以数组形式提供.如果为空则会返回所有的属性
			[]string{},
			nil,
		)

		sr, err := conn.Search(searchRequest)
		if err != nil {
			return nil, e.New(e.ValidateError, err)
		}
		if len(sr.Entries) != 1 {
			return nil, e.New(e.UserNotExists, err)
		}

		username := sr.Entries[0].GetAttributeValue("uid")
		if sr.Entries[0].GetAttributeValue("displayName") != "" {
			username = sr.Entries[0].GetAttributeValue("displayName")
		}

		users = append(users, &models.User{
			Email: email,
			Name:  username,
			Phone: sr.Entries[0].GetAttributeValue("mobile"),
		})
	}
	return users, nil
}

func SearchLdapUsers(q string, count int) ([]resps.LdapUserResp, e.Error) {
	conn, er := connectLdap()
	if er != nil {
		if er.Code() == e.LdapNotExisted {
			return nil, nil
		}
		return nil, e.New(e.LdapConnectFailed, er)
	}
	defer closeLdap(conn)

	conf := configs.Get()
	// SearchFilter 需要内填入搜索条件，单个用括号包裹，例如 (objectClass=person)(!(userAccountControl=514))
	emailAttr := "*"
	if q != "" {
		emailAttr = "*" + q + "*"
	}
	seachFilter := fmt.Sprintf("(&%s(%s=%s))", conf.Ldap.SearchFilter, conf.Ldap.EmailAttribute, emailAttr)
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
			DN:          sr.DN,
			Email:       sr.GetAttributeValue(conf.Ldap.EmailAttribute),
			Uid:         sr.GetAttributeValue(conf.Ldap.AccountAttribute),
			DisplayName: sr.GetAttributeValue("displayName"),
		})
	}

	if count > 0 && len(results) > count {
		return results[:count], nil
	}

	return results, nil
}

func CreateOUOrg(tx *db.Session, m models.LdapOUOrg) (models.Id, e.Error) {
	// 判断ou是否存在
	var ouOrg models.LdapOUOrg
	err := tx.Model(&models.LdapOUOrg{}).Where(`org_id = ?`, m.OrgId).Where(`dn = ?`, m.DN).First(&ouOrg)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", e.New(e.DBError, err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = tx.Insert(&m)
	} else {
		m.Id = ouOrg.Id
		_, err = tx.Model(&ouOrg).Update(models.LdapOUOrg{Role: m.Role})
	}

	if err != nil {
		return "", e.New(e.DBError, err)
	}

	return m.Id, nil
}

func CreateLdapUserOrg(tx *db.Session, orgId models.Id, m models.User, role string) (models.Id, e.Error) {
	var err error
	// 判断 user 是否存在
	var user models.User
	err = tx.Model(&models.User{}).Where(`email = ?`, m.Email).First(&user)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", e.New(e.DBError, err)
	}

	// 用户不存在
	userId := user.Id
	if errors.Is(err, gorm.ErrRecordNotFound) {
		m.Id = models.NewId("u")
		err = tx.Insert(&m)
		if err != nil {
			_ = tx.Rollback()
			return "", e.New(e.DBError, err)
		}

		userId = m.Id
	}

	// 用户授权不存在
	var userOrg models.UserOrg
	err = tx.Model(&models.UserOrg{}).Where("user_id = ? and org_id = ?", userId, orgId).First(&userOrg)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", e.New(e.DBError, err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = tx.Insert(&models.UserOrg{
			UserId:     userId,
			OrgId:      orgId,
			Role:       role,
			IsFromLdap: false,
		})
	} else {
		_, err = tx.Model(&userOrg).Update(models.UserOrg{Role: role})
	}

	if err != nil {
		return "", e.New(e.DBError, err)
	}

	return userId, nil
}

func CreateOUProject(tx *db.Session, m models.LdapOUProject) (models.Id, e.Error) {
	// 判断ou是否存在
	var ouProject models.LdapOUProject
	err := tx.Model(&models.LdapOUProject{}).Where(`org_id = ?`, m.OrgId).Where(`dn = ?`, m.DN).Where(`project_id = ?`, m.ProjectId).First(&ouProject)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", e.New(e.DBError, err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = tx.Insert(&m)
	} else {
		m.Id = ouProject.Id
		_, err = tx.Model(&ouProject).Update(models.LdapOUProject{Role: m.Role})
	}

	if err != nil {
		return "", e.New(e.DBError, err)
	}

	return m.Id, nil
}

// GetLdapOUOrgByDN 根据dn检索所有的org关联角色
func GetLdapOUOrgByDN(tx *db.Session, dn string) ([]models.LdapOUOrg, e.Error) {
	var results = make([]models.LdapOUOrg, 0)
	err := tx.Model(&models.LdapOUOrg{}).Where(`dn = ?`, dn).Find(&results)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return results, nil
}

// GetLdapOUProjectByDN 根据dn检索所有的project关联角色
func GetLdapOUProjectByDN(tx *db.Session, dn string) ([]models.LdapOUProject, e.Error) {
	var results = make([]models.LdapOUProject, 0)
	err := tx.Model(&models.LdapOUProject{}).Where(`dn = ?`, dn).Find(&results)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return results, nil
}
