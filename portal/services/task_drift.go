package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func QueryTaskDrift(query *db.Session) *db.Session {
	query = query.Model(&models.TaskDrift{})
	query = query.Joins("inner join iac_task it on it.id = iac_task_drift.task_id").
		LazySelectAppend("iac_task_drift.*,it.status")
	return query
}

func QueryResourceDrift(query *db.Session) *db.Session {
	query = query.Model(&models.Resource{}).LazySelectAppend("iac_resource.*")
	query = query.Joins("inner join iac_resource_drift rd on iac_resource.id = rd.res_id").
		LazySelectAppend("rd.drift_detail")
	return query
}

// GetLastTaskDrift 查询最新的一条漂移检测结果
func GetLastTaskDrift(tx *db.Session, envId models.Id) (*models.TaskDriftInfo, e.Error) {
	query := QueryTaskDrift(tx)
	//query := tx.Model(&models.TaskDrift{})
	query = query.Where("iac_task_drift.env_id = ?", envId).Limit(1)
	td := models.TaskDriftInfo{}
	if err := query.First(&td); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return &td, nil
}
