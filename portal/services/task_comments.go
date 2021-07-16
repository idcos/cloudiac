package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func CreateTaskComment(tx *db.Session, taskComment models.TaskComment) (*models.TaskComment, e.Error) {
	if err := models.Create(tx, &taskComment); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.TaskAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &taskComment, nil
}

func SearchTaskComment(dbsess *db.Session, taskId models.Id) *db.Session {
	return dbsess.Table(models.TaskComment{}.TableName()).Where("task_id = ?", taskId).Order("created_at desc")
}
