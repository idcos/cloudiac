package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func GetTask(dbSess *db.Session, id models.Id) (*models.Task, error) {
	task := models.Task{}
	err := dbSess.Where("id = ?", id).First(&task)
	return &task, err
}
