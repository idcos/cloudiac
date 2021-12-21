// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"fmt"
	"strings"
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
	}
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

func GetUserDetailById(query *db.Session, userId models.Id) (*models.UserWithRoleResp, e.Error) {
	d := models.UserWithRoleResp{}
	table := models.User{}.TableName()
	if err := query.Model(&models.User{}).Where(fmt.Sprintf("%s.id = ?", table), userId).First(&d); err != nil {
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
		Exists()
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

func UserHasManageUserPerm() {}

func QueryWithOrgId(query *db.Session, orgId interface{}, tableName ...string) *db.Session {
	return QueryWithCond(query, "org_id", orgId, tableName...)
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
	for _, userProject := range userProjects {
		userProjectsMap[userProject.ProjectId] = &userProject
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
