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

func GetProjectIdsByVgId(dbSess *db.Session, vgId models.Id) ([]string, error) {
	ids := make([]string, 0)
	if err := dbSess.Model(models.VariableGroupProjectRel{}).
		Where("var_group_id = ?", vgId).
		Pluck("project_id", &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// GetProjectEnvStat 环境状态占比
func GetProjectEnvStat(tx *db.Session, projectId models.Id) ([]resps.ProjectEnvStatResp, e.Error) {
	/* sample sql:
	select
		if(task_status = '',
		status,
		task_status) as status,
		id,
		name,
		count(*) as count
	from
		iac_env t
	where
		archived = 0
		and project_id = 'p-c8gg9josm56injdlb86g'
	group by
		t.status, t.id
	*/

	type dbResult struct {
		Status string
		Id     models.Id
		Name   string
		Count  int
	}

	query := tx.Model(&models.Env{}).Select(`if(task_status = '', status, task_status) as status, id, name, count(*) as count`)
	query = query.Where("archived = ?", 0).Where("project_id = ?", projectId)
	query = query.Group("t.status, t.id")

	var dbResults []dbResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	var m = make(map[string][]dbResult)
	var mTotalCount = make(map[string]int)
	for _, result := range dbResults {
		if _, ok := m[result.Status]; !ok {
			m[result.Status] = make([]dbResult, 0)
			mTotalCount[result.Status] = 0
		}
		m[result.Status] = append(m[result.Status], result)
		mTotalCount[result.Status] += result.Count
	}

	var results = make([]resps.ProjectEnvStatResp, 0)
	for k, v := range m {
		data := resps.ProjectEnvStatResp{
			Status: k,
			Count:  mTotalCount[k],
			Envs:   make([]resps.EnvDetailStatResp, 0),
		}

		for _, e := range v {
			data.Envs = append(data.Envs, resps.EnvDetailStatResp{
				Id:    e.Id,
				Name:  e.Name,
				Count: e.Count,
			})
		}
		results = append(results, data)
	}

	return results, nil
}

// GetProjectResStat 资源类型占比
func GetProjectResStat(tx *db.Session, projectId models.Id, limit int) ([]resps.EnvResStatResp, e.Error) {
	/* sample sql
	select
		iac_resource.type as res_type,
		iac_env.id as id,
		iac_env.name as name,
		count(*) as count
	from
		iac_resource
	join iac_env on
		iac_env.last_res_task_id = iac_resource.task_id
		and iac_env.id = iac_resource.env_id
	where
		iac_env.project_id = 'p-c8gg9josm56injdlb86g'
	group by
		iac_resource.type,
		iac_env.id
	order by
		count desc
	limit 10;
	*/

	query := tx.Model(&models.Resource{}).Select(`iac_resource.type as res_type, iac_env.id as id, iac_env.name as name, count(*) as count`)
	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id and iac_env.id = iac_resource.env_id`)
	query = query.Where(`iac_env.project_id = ?`, projectId)

	query = query.Group("iac_resource.type, iac_env.id").Order("count desc")
	if limit > 0 {
		query = query.Limit(limit)
	}

	type dbResult struct {
		ResType string
		Id      models.Id
		Name    string
		Count   int
	}

	var dbResults []dbResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	var m = make(map[string][]dbResult)
	var mTotalCount = make(map[string]int)
	for _, result := range dbResults {
		if _, ok := m[result.ResType]; !ok {
			m[result.ResType] = make([]dbResult, 0)
			mTotalCount[result.ResType] = 0
		}
		m[result.ResType] = append(m[result.ResType], result)
		mTotalCount[result.ResType] += result.Count
	}

	var results []resps.EnvResStatResp
	for k, v := range m {
		data := resps.EnvResStatResp{
			ResType: k,
			Count:   mTotalCount[k],
			Envs:    make([]resps.EnvDetailStatResp, 0),
		}

		for _, e := range v {
			data.Envs = append(data.Envs, resps.EnvDetailStatResp{
				Id:    e.Id,
				Name:  e.Name,
				Count: e.Count,
			})
		}
		results = append(results, data)
	}

	return results, nil
}

// GetProjectEnvResStat 环境资源数量
func GetProjectEnvResStat(tx *db.Session, projectId models.Id, limit int) ([]resps.EnvResStatResp, e.Error) {

	query := tx.Model(&models.Resource{}).Select(`iac_resource.env_id as env_id, iac_env.name as env_name, iac_resource.type as res_type, DATE_FORMAT(iac_resource.applied_at, "%Y-%m") as date, count(*) as count`)

	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id`)
	query = query.Where(`iac_env.project_id = ?`, projectId)
	query = query.Where(`DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(CURDATE(), "%Y-%m") OR DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), "%Y-%m")`)

	query = query.Group("iac_resource.type,iac_resource.env_id,date")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var results []resps.EnvResStatResp
	if err := query.Find(&results); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return results, nil
}

// GetProjectResGrowTrend 最近7天资源及费用趋势
func GetProjectResGrowTrend(tx *db.Session, projectId models.Id, days int) ([]resps.ResGrowTrendResp, e.Error) {

	query := tx.Model(&models.Resource{}).Select(`DATE_FORMAT(applied_at, "%Y-%m-%d") as date, count(*) as count`)
	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id`)

	query = query.Where("iac_env.project_id = ?", projectId)

	query = query.Where(`applied_at > DATE_SUB(CURDATE(), INTERVAL ? DAY) or (applied_at > DATE_SUB(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), INTERVAL ? DAY) and applied_at <= DATE_SUB(CURDATE(), INTERVAL 1 MONTH))`, days, days)

	query = query.Group("date").Order("date")

	var results []resps.ResGrowTrendResp
	if err := query.Find(&results); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return nil, nil
}
