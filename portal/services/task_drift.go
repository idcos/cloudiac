package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func QueryTaskDrift(query *db.Session) *db.Session {
	query = query.Model(&models.TaskDrift{})
	query.Joins("inner join iac_task it on it.id = iac_task_drift.task_id").
		LazySelectAppend("it.status")
	return query
}

func QueryResourceDrift(query *db.Session) *db.Session {
	query = query.Model(&models.Resource{})
	query = query.Joins("inner join iac_resource_drift rd on iac_resource.id = rd.res_id").
		LazySelectAppend("rd.drift_detail")
	return query
}
