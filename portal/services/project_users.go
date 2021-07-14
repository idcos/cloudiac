package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/utils"
)

func BindProjectUsers(tx *db.Session, projectId models.Id, authorization []forms.UserAuthorization) e.Error {
	bq := utils.NewBatchSQL(1024, "INSERT INTO", models.UserProject{}.TableName(),
		"user_id", "project_id", "role")

	for _, v := range authorization {
		if err := bq.AddRow(v.UserId, projectId, v.Role); err != nil {
			return e.New(e.DBError, err)
		}
	}

	for bq.HasNext() {
		sql, args := bq.Next()
		if _, err := tx.Exec(sql, args...); err != nil {
			return e.New(e.DBError, err)
		}
	}
	return nil
}

func UpdateProjectUsers(tx *db.Session, projectId models.Id, authorization []forms.UserAuthorization) e.Error {
	if _, err := tx.Where("project_id = ?", projectId).Delete(&models.UserProject{}); err != nil {
		return e.New(e.DBError, err)
	}
	return BindProjectUsers(tx, projectId, authorization)
}

func SearchProjectUsers(tx *db.Session, projectId models.Id) ([]models.UserProject, e.Error) {
	userProject := make([]models.UserProject, 0)
	if err := tx.Where("project_id = ?", projectId).Find(&userProject); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return userProject, nil
}

// GetProjectsByUser 获取用户关联的所有 project id
func GetProjectsByUser(tx *db.Session, userId models.Id) ([]models.Id, e.Error) {
	ids := make([]models.Id, 0)
	if err := tx.Model(models.UserProject{}).Where("user_id = ?", userId).Find(&ids); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return ids, nil
}

// GetProjectsByUserOrg 获取用户在该组织下的 project id
func GetProjectsByUserOrg(tx *db.Session, userId models.Id, orgId models.Id) ([]models.Id, e.Error) {
	ids := make([]models.Id, 0)
	if err := tx.Model(models.Project{}).Where("org_id = ?", orgId).
		Joins("left join %s as o on o.project_id = %s.id where o.user_id = ?",
			models.UserProject{}.TableName(), models.Project{}.TableName(), userId).
		Find(&ids); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return ids, nil
}

// GetProjectsByOrg 获取组织下的所有 project id
func GetProjectsByOrg(tx *db.Session, orgId models.Id) ([]models.Id, e.Error) {
	ids := make([]models.Id, 0)
	if err := tx.Model(models.Project{}).Where("org_id = ?", orgId).Find(&ids); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return ids, nil
}

// GetProjectRoleByUser 获取用户在项目中的角色
func GetProjectRoleByUser(tx *db.Session, projectId models.Id, userId models.Id) (string, e.Error) {
	var role string
	if err := tx.Model(models.UserProject{}).Where("user_id = ? AND project_id = ?", projectId, userId).Find(&role); err != nil {
		return "", e.AutoNew(err, e.DBError)
	}
	return role, nil
}
