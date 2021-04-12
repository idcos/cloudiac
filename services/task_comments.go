package services

import (
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
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

func SearchTaskComment(dbsess *db.Session, taskId uint) *db.Session {
	return dbsess.Table(models.TaskComment{}.TableName()).Where("task_id = ?", taskId).Order("created_at desc")
}
