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
	query := dbSess.Table(models.Project{}.TableName()).Where("org_id = ?",orgId)
	if q != "" {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", q))
	}
	return query.Order("created_at DESC")
}

func DeleteUserProject(tx *db.Session) {

}