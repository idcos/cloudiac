package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func GetTaskSteps(sess *db.Session, taskId models.Id) ([]*models.TaskStep, error) {
	steps := make([]*models.TaskStep, 0)
	err := sess.Where(models.TaskStep{TaskId: taskId}).Order("index").Find(&steps)
	return steps, err
}

func GetTaskStep(sess *db.Session, taskId models.Id, step int) (*models.TaskStep, error) {
	taskStep := models.TaskStep{}
	err := sess.Where(models.TaskStep{
		TaskId: taskId,
		Index:  step,
	}).First(&taskStep)
	return &taskStep, err
}
