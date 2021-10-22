package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func GetTaskJob(sess *db.Session, jobId models.Id) (*models.TaskJob, e.Error) {
	job := models.TaskJob{}
	err := sess.Model(&models.TaskJob{}).Where("id = ?", jobId).First(&job)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.AutoNew(err, e.ObjectNotExists)
		}
		return nil, e.AutoNew(err, e.DBError)
	}
	return &job, nil
}

func GetTaskJobs(sess *db.Session, taskId models.Id) (rs []models.TaskJob, er e.Error) {
	err := sess.Model(&models.TaskJob{}).Where("task_id = ?", taskId).Find(&rs)
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return rs, nil
}

func UpdateTaskJobContainerId(sess *db.Session, jobId models.Id, containerId string) e.Error {
	_, err := models.UpdateModel(
		sess,
		&models.TaskJob{ContainerId: containerId},
		"id = ?", jobId,
	)
	if err != nil {
		return e.AutoNew(err, e.DBError)
	}
	return nil
}
