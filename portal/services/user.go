// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/resps"
	"cloudiac/utils"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

func CreateUser(tx *db.Session, user models.User) (*models.User, e.Error) {
	if user.Id == "" {
		user.Id = models.NewId("u")
	}
	if err := models.Create(tx, &user); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.UserAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &user, nil
}

func UpdateUser(tx *db.Session, id models.Id, attrs models.Attrs) (user *models.User, re e.Error) {
	user = &models.User{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.User{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.UserEmailDuplicate)
		}
		return nil, e.New(e.DBError, fmt.Errorf("update user error: %v", err))
	} //nolint
	if err := tx.Where("id = ?", id).First(user); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query user error: %v", err))
	}
	return
}

func DeleteUser(tx *db.Session, id models.Id) e.Error {
	if _, err := tx.Where("id = ?", id).Delete(&models.User{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete user error: %v", err))
	}
	return nil
}

// GetUserByIdRaw 按 ID 查找用户，不排除系统用户
func GetUserByIdRaw(tx *db.Session, id models.Id) (*models.User, e.Error) {
	u := models.User{}
	if err := tx.Where("id = ?", id).First(&u); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.UserNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &u, nil
}

// RefreshUserOrgRoles 刷新用户的组织权限
func RefreshUserOrgRoles(tx *db.Session, userId models.Id, ldapUserOrgOUs []models.LdapOUOrg) e.Error {
	_, err := tx.Where(`user_id = ?`, userId).Delete(&models.UserOrg{})
	if err != nil {
		return e.New(e.DBError, err)
	}

	userOrgs := make([]models.UserOrg, 0)
	for _, item := range ldapUserOrgOUs {
		userOrgs = append(userOrgs, models.UserOrg{
			UserId: userId,
			OrgId:  item.OrgId,
			Role:   item.Role,
		})
	}

	err = tx.Insert(&userOrgs)
	if err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

// RefreshUserProjectRoles 刷新用户的项目权限
func RefreshUserProjectRoles(tx *db.Session, userId models.Id, ldapUserProjectOUs []models.LdapOUProject) e.Error {
	_, err := tx.Where(`user_id = ?`, userId).Delete(&models.UserProject{})
	if err != nil {
		return e.New(e.DBError, err)
	}

	userProjects := make([]models.UserProject, 0)
	for _, item := range ldapUserProjectOUs {
		userProjects = append(userProjects, models.UserProject{
			UserId:    userId,
			ProjectId: item.ProjectId,
			Role:      item.Role,
		})
	}
	err = tx.Insert(&userProjects)
	if err != nil {
		return e.New(e.DBError, err)
	}

	return nil
}

// GetUserById 按 ID 查找用户
func GetUserById(tx *db.Session, id models.Id) (*models.User, e.Error) {
	tx = tx.Where("id != ?", consts.SysUserId)
	return GetUserByIdRaw(tx, id)
}

func GetUserByEmail(tx *db.Session, email string) (*models.User, e.Error) {
	u := models.User{}
	if err := tx.Where("email = ?", email).First(&u); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.UserNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &u, nil
}

func FindUsers(query *db.Session) (users []*models.User, err error) {
	err = query.Find(&users)
	return
}

func QueryUser(query *db.Session) *db.Session {
	return query.Model(&models.User{})
}

func HashPassword(password string) (string, e.Error) {
	if er := CheckPasswordFormat(password); er != nil {
		return "", er
	}

	hashed, err := utils.HashPassword(password)
	if err != nil {
		return "", e.New(e.InternalError, err)
	}
	return hashed, nil
}

func CheckPasswordFormat(password string) e.Error {
	//密码校验规则：数字、大写字母、小写字母两种及两种以上组合，且长度在6~30之间
	if len(password) < 6 || len(password) > 30 {
		return e.New(e.InvalidPasswordFormat, fmt.Errorf("error: password length"))
	}

	typeCount := 0
	for _, chars := range []string{consts.LowerCaseLetter, consts.UpperCaseLetter, consts.DigitChars} {
		if strings.ContainsAny(password, chars) {
			typeCount += 1
		}
	}
	if typeCount < 2 {
		return e.New(e.InvalidPasswordFormat, fmt.Errorf("error: password strength"))
	}

	return nil
}

func GetUserDetailById(query *db.Session, userId models.Id) (*resps.UserWithRoleResp, e.Error) {
	d := resps.UserWithRoleResp{}
	table := models.User{}.TableName()
	if err := query.Model(&models.User{}).
		Where(fmt.Sprintf("%s.id = ?", table), userId).
		LazySelectAppend("iac_user.*").
		Scan(&d); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.UserNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &d, nil
}

func GetUserRoleByOrg(dbSess *db.Session, userId, orgId models.Id, role string) (bool, e.Error) {
	isExists, err := dbSess.Table(models.UserOrg{}.TableName()).
		Where("user_id = ?", userId).
		Where("role = ?", role).
		Where("org_id = ?", orgId).
		Exists() //nolint
	if err != nil {
		return isExists, e.New(e.DBError, err)
	}
	return isExists, nil
}

func GetUserRoleByProject(dbSess *db.Session, userId, projectId models.Id, role string) (bool, e.Error) {
	isExists, err := dbSess.Table(models.UserProject{}.TableName()).
		Where("user_id = ?", userId).
		Where("role = ?", role).
		Where("project_id = ?", projectId).
		Exists()
	if err != nil {
		return isExists, e.New(e.DBError, err)
	}
	return isExists, nil
}

// ========================================================================

// UserOrgIds 用户关联组织 id 列表
func UserOrgIds(userId models.Id) []models.Id {
	var ids []models.Id
	userOrgs := getUserOrgs(userId)
	for _, userOrg := range userOrgs {
		ids = append(ids, userOrg.OrgId)
	}
	return ids
}

// UserProjectIds 用户有权限的项目 id
func UserProjectIds(userId models.Id, orgId models.Id) []models.Id {
	var ids []models.Id
	userProjects := getUserProjects(userId)
	for _, userProject := range userProjects {
		ids = append(ids, userProject.ProjectId)
	}
	return ids
}

// UserOrgRoles 用户在组织(多个)下的角色
// @return map[models.Id]*models.UserOrg 返回 map[orgId]UserOrg
func UserOrgRoles(userId models.Id) map[models.Id]*models.UserOrg {
	return getUserOrgs(userId)
}

// UserProjectRoles 用户在项目(多个)下的角色
// @return map[models.Id]*models.UserProject 返回 map[projectId]UserProject
func UserProjectRoles(userId models.Id) map[models.Id]*models.UserProject {
	return getUserProjects(userId)
}

// UserHasProjectRole 用户是否拥有项目的某个角色权限
func UserHasProjectRole(userId models.Id, orgId models.Id, projectId models.Id, role string) bool {
	// TODO 临时处理系统管理员权限
	if userId.String() == consts.SysUserId {
		return true
	}

	// 用户属于项目?
	userProjects := getUserProjects(userId)
	if userProjects[projectId] == nil {
		return false
	}

	if role == "" {
		return true
	} else {
		return role == userProjects[projectId].Role
	}
}

// UserHasOrgRole 用户是否拥有组织的某个角色权限
// role 传空字符串表示只检查 user 是否属于 org，不检查具体 role
func UserHasOrgRole(userId models.Id, orgId models.Id, role string) bool {
	// TODO 临时处理系统管理员权限
	if userId.String() == consts.SysUserId {
		return true
	}
	userOrgs := getUserOrgs(userId)
	if userOrgs[orgId] == nil {
		return false
	}
	if role == "" {
		return true
	} else {
		return role == userOrgs[orgId].Role
	}
}

// UserIsSuperAdmin 判断用户是否是平台管理员
func UserIsSuperAdmin(query *db.Session, userId models.Id) bool {
	if user, err := GetUserById(query, userId); err != nil {
		return false
	} else {
		return user.IsAdmin
	}
}

func QueryWithOrgId(query *db.Session, orgId interface{}, tableName ...string) *db.Session {
	return QueryWithCond(query, "org_id", orgId, tableName...)
}

func QueryWithOrgIdAndGlobal(query *db.Session, orgId interface{}, tableName ...string) *db.Session {
	if len(tableName) > 0 {
		return query.Where(fmt.Sprintf("`%s`.`org_id` = ? or `%s`.`org_id` = ''", tableName[0], orgId))
	}
	return query.Where("`org_id` = ? or `org_id` = ''", orgId)
}

func QueryWithProjectId(query *db.Session, projectId interface{}, tableName ...string) *db.Session {
	return QueryWithCond(query, "project_id", projectId, tableName...)
}

func QueryWithOrgProject(query *db.Session, orgId interface{}, projId interface{}, tableName ...string) *db.Session {
	return QueryWithProjectId(QueryWithOrgId(query, orgId, tableName...), projId, tableName...)
}

func QueryWithCond(query *db.Session, column string, value interface{}, tableName ...string) *db.Session {
	if len(tableName) > 0 {
		return query.Where(fmt.Sprintf("`%s`.`%s` = ?", tableName[0], column), value)
	}
	return query.Where(fmt.Sprintf("`%s` = ?", column), value)
}

// TODO lru cache data
// getUserOrgs 获取用户组织关联列表
// @return map[models.Id]*models.UserOrg 返回 map[orgId]UserOrg
func getUserOrgs(userId models.Id) map[models.Id]*models.UserOrg {
	userOrgs := make([]models.UserOrg, 0)
	query := db.Get()
	if err := query.Model(models.UserOrg{}).Where("user_id = ?", userId).Find(&userOrgs); err != nil {
		return nil
	}
	userOrgsMap := make(map[models.Id]*models.UserOrg)
	for index, userOrg := range userOrgs {
		userOrgsMap[userOrg.OrgId] = &userOrgs[index]
	}
	return userOrgsMap
}

// getUserProjects 获取用户项目关联列表
// @return map[models.Id]*models.UserProject 返回 map[projectId]UserProject
func getUserProjects(userId models.Id) map[models.Id]*models.UserProject {
	userProjects := make([]models.UserProject, 0)
	query := db.Get()
	if err := query.Model(models.UserProject{}).Where("user_id = ?", userId).Find(&userProjects); err != nil {
		return nil
	}
	userProjectsMap := make(map[models.Id]*models.UserProject)
	for index, userProject := range userProjects {
		userProjectsMap[userProject.ProjectId] = &userProjects[index]
	}
	return userProjectsMap
}

func GetUsersByUserIds(dbSess *db.Session, userId []string) []models.User {
	users := make([]models.User, 0)
	if err := dbSess.Where("id in (?)", userId).Find(users); err != nil {
		return nil
	}
	return users
}

// 处理Ldap 登录逻辑
func LdapAuthLogin(userEmail, password string) (username, dn string, er e.Error) {
	conf := configs.Get()
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", conf.Ldap.LdapServer, conf.Ldap.LdapServerPort))
	if err != nil {
		return username, dn, e.New(e.LdapConnectFailed, err)
	}
	defer conn.Close()
	// 配置ldap 管理员dn信息，例如cn=Manager,dc=idcos,dc=com
	err = conn.Bind(conf.Ldap.AdminDn, conf.Ldap.AdminPassword)
	if err != nil {
		return username, dn, e.New(e.ValidateError, err)
	}
	// SearchFilter 需要内填入搜索条件，单个用括号包裹，例如 (objectClass=person)(!(userAccountControl=514))
	seachFilter := fmt.Sprintf("(&%s(%s=%s))", conf.Ldap.SearchFilter, conf.Ldap.EmailAttribute, userEmail)
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
		return username, dn, e.New(e.ValidateError, err)
	}
	if len(sr.Entries) != 1 {
		return username, dn, e.New(e.UserNotExists, err)
	}
	err = conn.Bind(sr.Entries[0].DN, password)
	if err != nil {
		return username, dn, e.New(e.InvalidPassword, err)
	}
	var account string
	if conf.Ldap.AccountAttribute != "" {
		account = conf.Ldap.AccountAttribute
	} else {
		account = "uid"
	}
	return sr.Entries[0].GetAttributeValue(account), sr.Entries[0].DN, nil

}

// GetUserHighestProjectRole 获取用户在指定组织下的最高项目角色
func GetUserHighestProjectRole(db *db.Session, orgId models.Id, userId models.Id) (string, e.Error) {
	roles := make([]string, 0)
	err := db.Model(&models.UserProject{}).
		Joins("join iac_project as p on p.id = iac_user_project.project_id and p.org_id = ?", orgId).
		Where("user_id = ?", userId).Group("role").
		Select("role").Scan(&roles)
	if err != nil {
		return "", e.New(e.DBError, err)
	}
	return GetHighestRole(roles...), nil
}

// HasInviteUserPerm 判断用户是否有邀请其他用户加入组织的权限
func HasInviteUserPerm(db *db.Session, userId models.Id, orgId models.Id, targetRole string) (bool, e.Error) {
	if UserHasOrgRole(userId, orgId, consts.OrgRoleAdmin) {
		return true, nil
	}

	if targetRole == consts.OrgRoleMember {
		// 组织下的项目管理员可以邀请用户成为组织成员
		role, err := GetUserHighestProjectRole(db, orgId, userId)
		if err != nil {
			return false, err
		} else if role == consts.ProjectRoleManager {
			return true, nil
		}
	}
	return false, nil
}
