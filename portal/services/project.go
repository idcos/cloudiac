// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/resps"
	"fmt"
)

func CreateProject(tx *db.Session, project *models.Project) (*models.Project, e.Error) {
	if project.Id == "" {
		project.Id = models.NewId("p")
	}
	if err := models.Create(tx, project); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.ProjectAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return project, nil
}

func SearchProject(dbSess *db.Session, orgId models.Id, q, status string) *db.Session {
	query := dbSess.Model(&models.Project{}).Where(fmt.Sprintf("%s.org_id = ?", models.Project{}.TableName()), orgId)
	if q != "" {
		query = query.Where(fmt.Sprintf("%s.name like ?", models.Project{}.TableName()), fmt.Sprintf("%%%s%%", q))
	}
	if status != "" {
		query = query.Where(fmt.Sprintf("%s.`status` = ?", models.Project{}.TableName()), status)
	}
	return query
}

func UpdateProject(tx *db.Session, project *models.Project, attrs map[string]interface{}) e.Error {
	if _, err := models.UpdateAttr(tx, project, attrs); err != nil {
		if e.IsDuplicate(err) {
			return e.New(e.ProjectAliasDuplicate)
		}
		return e.New(e.DBError, err)
	}
	return nil
}

func DetailProject(dbSess *db.Session, projectId models.Id) (models.Project, e.Error) {
	project := models.Project{}
	if err := dbSess.Where("id = ?", projectId).First(&project); err != nil {
		return project, e.New(e.DBError, err)
	}
	return project, nil
}

func DeleteProject(tx *db.Session, projectId models.Id) e.Error {
	if _, err := tx.Where("id = ?", projectId).Delete(&models.Project{}); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

// StatisticalProjectTpl todo 项目统计 待完善
func StatisticalProjectTpl(dbSess *db.Session, projectId models.Id) (int64, error) {
	return dbSess.Table(models.ProjectTemplate{}.TableName()).Where("project_id = ?", projectId).Count()
}

func StatisticalProjectEnv(dbSess *db.Session, projectId models.Id) (*struct {
	EnvActive   int64
	EnvFailed   int64
	EnvInactive int64
}, error) {
	var (
		resp []struct {
			Count  int64
			Status string
		}
		envActive   int64
		envFailed   int64
		envInactive int64
	)

	if err := dbSess.Model(&models.Env{}).Select("count(status) as count, status").
		Where("project_id = ?", projectId).Group("status").Find(&resp); err != nil {
		return nil, err
	}

	for _, v := range resp {
		switch v.Status {
		case models.EnvStatusFailed:
			envFailed = v.Count
		case models.EnvStatusActive:
			envActive = v.Count
		case models.EnvStatusInactive:
			envInactive = v.Count
		}
	}

	return &struct {
		EnvActive   int64
		EnvFailed   int64
		EnvInactive int64
	}{
		EnvActive:   envActive,
		EnvFailed:   envFailed,
		EnvInactive: envInactive,
	}, nil

}

func GetProjectEnvStat(tx *db.Session, projectId models.Id) ([]resps.EnvStatResp, e.Error) {
	subQuery := tx.Model(&models.Env{}).Select(`if(task_status = '', status, task_status) as status`)
	subQuery = subQuery.Where("archived = ?", 0).Where("project_id = ?", projectId)

	query := tx.Table("(?) as t", subQuery.Expr()).Select(`status, count(*) as count`).Group("status")

	var results []resps.EnvStatResp
	if err := query.Find(&results); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return results, nil
}

func GetProjectResStat(tx *db.Session, projectId models.Id, limit int) ([]resps.ResStatResp, e.Error) {
	query := tx.Model(&models.Resource{}).Select(`iac_resource.type as res_type, count(*) as count`)
	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id`)
	query = query.Where(`iac_env.project_id = ?`, projectId)

	query = query.Group("res_type").Order("count desc")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var results []resps.ResStatResp
	if err := query.Find(&results); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return results, nil
}

func GetProjectEnvResStat(tx *db.Session, projectId models.Id, limit int) ([]resps.EnvResStatResp, e.Error) {

	return nil, nil
}

func GetProjectResGrowTrend(tx *db.Session, projectId models.Id, days int) ([]resps.ResGrowTrendResp, e.Error) {

	return nil, nil
}

// GetProjectResSummary 项目资源概览
func GetProjectResSummary(tx *db.Session, projectId models.Id, limit int) ([]resps.EnvResSummaryResp, e.Error) {

	return nil, nil
}
