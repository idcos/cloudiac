package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func QueryTaskDrift(query *db.Session) *db.Session {
	query = query.Model(&models.TaskDrift{})
	return query
}
