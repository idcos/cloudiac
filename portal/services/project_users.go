package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/utils"
	"fmt"
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
	if err := tx.Model(models.UserProject{}).Where("user_id = ?", userId).Pluck("id", &ids); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return ids, nil
}

// GetProjectsByUserOrg 获取用户在该组织下的 project id
func GetProjectsByUserOrg(tx *db.Session, userId models.Id, orgId models.Id) ([]models.Id, e.Error) {
	ids := make([]models.Id, 0)
	if err := tx.Model(models.Project{}).
		Joins(fmt.Sprintf("left join %s as o on o.project_id = %s.id",
			models.UserProject{}.TableName(), models.Project{}.TableName())).
		LazySelectAppend(fmt.Sprintf("o.*,%s.*", models.Project{}.TableName())).
		Where(fmt.Sprintf("o.user_id = ? and %s.org_id = ?", models.Project{}.TableName()), userId, orgId).
		Pluck(fmt.Sprintf("%s.id", models.Project{}.TableName()), &ids); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return ids, nil
}

// GetProjectsByOrg 获取组织下的所有 project id
func GetProjectsByOrg(tx *db.Session, orgId models.Id) ([]models.Id, e.Error) {
	ids := make([]models.Id, 0)
	if err := tx.Model(models.Project{}).Where("org_id = ?", orgId).
		Pluck(fmt.Sprintf("%s.id", models.Project{}.TableName()), &ids); err != nil {
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

// GetProjectsById 获取项目
func GetProjectsById(tx *db.Session, projectId models.Id) (*models.Project, e.Error) {
	proj := models.Project{}
	if err := tx.Model(models.Project{}).Where("id = ?", projectId).First(&proj); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return &proj, nil
}

func GetUserIdsByProject(query *db.Session, projectId models.Id) ([]models.Id, e.Error) {
	var userProjects []*models.UserProject
	var userIds []models.Id
	if err := query.Where("project_id = ?", projectId).Find(&userProjects); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	for _, p := range userProjects {
		userIds = append(userIds, p.UserId)
	}
	return userIds, nil
}

//GetUserIdsByProjectUser 获取项目下的userid
func GetUserIdsByProjectUser(query *db.Session, projectId models.Id) ([]models.Id, e.Error) {
	var userIds []models.Id
	if err := query.Table(models.UserProject{}.TableName()).
		Where("project_id = ?", projectId).
		Pluck("user_id", &userIds); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return userIds, nil
}

func CreateProjectUser(dbSess *db.Session, userProject models.UserProject) (*models.UserProject, e.Error) {
	if err := models.Create(dbSess, &userProject); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.ProjectUserAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &userProject, nil
}

func UpdateProjectUser(dbSess *db.Session, attrs models.Attrs) e.Error {
	if _, err := models.UpdateAttr(dbSess, &models.UserProject{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return e.New(e.ProjectUserAliasDuplicate)
		}
		return e.New(e.DBError, fmt.Errorf("update project user error: %v", err))
	}
	return nil
}

func DeleteProjectUser(dbSess *db.Session, id uint) e.Error {
	if _, err := dbSess.Where("id = ?", id).Delete(&models.UserProject{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete project error: %v", err))
	}
	return nil
}

// GetDemoProject 获取演示项目
func GetDemoProject(tx *db.Session, demoOrgId models.Id) (*models.Project, e.Error) {
	proj := models.Project{}
	if err := tx.Model(models.Project{}).
		Joins("left join iac_user_project on iac_user_project.project_id = iac_project.id").
		Where("org_id = ?", demoOrgId).First(&proj); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return &proj, nil
}
