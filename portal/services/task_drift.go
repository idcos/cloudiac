package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func QueryTaskDrift(query *db.Session) *db.Session {
	query = query.Model(&models.TaskDrift{})
	return query
}

func QueryResourceDrift(query *db.Session) *db.Session {
	query = query.Model(&models.Resource{})
	query = query.Joins("inner join iac_resource_drift rd on iac_resource.id = rd.res_id").
		LazySelectAppend("iac_resource.*,res.provider as provider,res.")
	return query
}
