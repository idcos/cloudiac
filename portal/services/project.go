package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

func CreateProject(tx *db.Session, project *models.Project) (*models.Project, e.Error) {
	if err := models.Create(tx, project); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return project, nil
}

func CreateUserProject(tx *db.Session, userProjects []models.Modeler) e.Error {
	if err := models.CreateBatch(tx, userProjects); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

func SearchProject(dbSess *db.Session, orgId models.Id, q string) *db.Session {
	query := dbSess.Table(models.Project{}.TableName()).Where("org_id = ?", orgId)
	if q != "" {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", q))
	}
	return query.Order("created_at DESC")
}

func DeleteUserProject(tx *db.Session, projectId models.Id) e.Error {
	if _, err := tx.
		Where("project_id = ?", projectId).
		Delete(models.UserProject{}); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

func UpdateProject(tx *db.Session, project *models.Project, attrs map[string]interface{}) e.Error {
	if _, err := models.UpdateAttr(tx, project, attrs); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

func DetailProject(dbSess *db.Session, projectId models.Id) (interface{}, e.Error) {
	project := models.Project{}
	if err := dbSess.Where("id = ?", projectId).First(&project); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return project, nil
}
