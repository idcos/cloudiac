package services

import (
	"fmt"
	"strings"

	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
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

func GetUserById(tx *db.Session, id models.Id) (*models.User, e.Error) {
	u := models.User{}
	if err := tx.Where("id = ?", id).First(&u); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.UserNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &u, nil
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

// UserOrgIds 用户有权限的组织 id
func UserOrgIds(query *db.Session, userId models.Id) (*[]models.Id, e.Error) {
	ids := make([]models.Id, 0)
	if UserIsSuperAdmin(query, userId) {
		// 平台管理员允许访问所有组织
		err := query.Model(models.Organization{}).Find(&ids)
		if err != nil && !e.IsRecordNotFound(err) {
			return nil, e.New(e.DBError, err)
		}
	} else {
		// 其他用户只允许访问关联的组织
		err := query.Model(models.UserOrg{}).Where("user_id = ?", userId).Find(&ids)
		if err != nil && !e.IsRecordNotFound(err) {
			return nil, e.New(e.DBError, err)
		}
	}
	return &ids, nil
}

// UserProjectIds 用户有权限的项目 id
func UserProjectIds(query *db.Session, userId models.Id, orgId models.Id) (*[]models.Id, e.Error) {
	ids := make([]models.Id, 0)
	user, _ := GetUserById(query, userId)
	if user == nil {
		return &ids, nil
	}
	if UserIsSuperAdmin(query, userId) {
		// 平台管理员允许访问所有项目
		err := query.Model(models.Project{}).Find(&ids)
		if err != nil && !e.IsRecordNotFound(err) {
			return nil, e.New(e.DBError, err)
		}
	} else {
		// 其他用户只允许访问关联的组织
		err := query.Model(models.UserOrg{}).Where("user_id = ?", userId).Find(&ids)
		if err != nil && !e.IsRecordNotFound(err) {
			return nil, e.New(e.DBError, err)
		}
	}
	return &ids, nil
}

// UserOrgRoles 用户在组织(多个)下的角色
func UserOrgRoles(query *db.Session, userId models.Id, orgId models.Id) {

}

// UserProjectRoles 用户在项目(多个)下的角色
func UserProjectRoles() {

}

// UserHasProjectRole 用户是否拥有项目的某个角色权限
func UserHasProjectRole(query *db.Session, userId models.Id, orgId models.Id, projectId models.Id, role ...string) bool {
	switch {
	case UserIsSuperAdmin(query, userId):
		// 平台管理员拥有所有权限
		return true
	case UserHasOrgRole(query, userId, orgId, consts.OrgRoleAdmin):
		// 组织管理员拥有组织下项目的管理者权限
		if exist, err := query.Model(models.Project{}).Where("org_id = ?", orgId).Exists(); err != nil {
			return false
		} else {
			if !exist {
				return false
			}
			if len(role) == 0 {
				// 不关心具体角色，只检查是否关联了项目
				return true
			} else {
				// 是否有确定的项目角色
				return role[0] == consts.ProjectRoleManager
			}
		}
	default:
		// 普通用户
		r := models.UserProject{}
		err := query.Model(models.UserProject{}).Where("project_id = ? and user_id = ?", projectId, userId).Find(&r)
		if err != nil {
			return false
		} else if len(role) == 0 {
			// 不关心具体角色，只检查是否关联了项目
			return true
		} else {
			// 是否有确定的项目角色
			return r.Role == role[0]
		}
	}
}

// UserHasOrgRole 用户是否拥有组织的某个角色权限
func UserHasOrgRole(query *db.Session, userId models.Id, orgId models.Id, role ...string) bool {
	if UserIsSuperAdmin(query, userId) {
		// 平台管理员拥有所有权限
		return true
	}
	ur := models.UserOrg{}
	err := query.Model(models.UserOrg{}).Where("user_id = ? AND org_id = ?", userId, orgId).Find(&ur)
	if err != nil {
		return false
	} else if len(role) == 0 {
		// 不关心具体角色，只检查是否关联了组织
		return true
	} else {
		// 是否有确定的组织角色
		return role[0] == ur.Role
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

func QueryWithCond(query *db.Session, column string, value interface{}, tableName ...string) *db.Session {
	if len(tableName) > 0 {
		return query.Where(fmt.Sprintf("`%s`.`%s` = ?", tableName[0], column), value)
	}
	return query.Where(fmt.Sprintf("`%s` = ?", column), value)
}

// TODO lru cache data
//userOrgs	 = map[string]*models.UserOrg
//userProjects = map[string]*models.UserProject
