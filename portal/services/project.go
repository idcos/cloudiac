// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/resps"
	"fmt"
	"time"
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
func GetProjectEnvStat(tx *db.Session, projectId models.Id) ([]resps.EnvStatResp, e.Error) {
	/* sample sql:
	select
		if(task_status = '',
		status,
		task_status) as my_status,
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
		MyStatus string
		Id       models.Id
		Name     string
		Count    int
	}

	query := tx.Model(&models.Env{}).Select(`if(task_status = '', status, task_status) as my_status, id, name, count(*) as count`)
	query = query.Where("archived = ?", 0).Where("project_id = ?", projectId)
	query = query.Group("my_status, id")

	var dbResults []dbResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	var m = make(map[string][]dbResult)
	var mTotalCount = make(map[string]int)
	for _, result := range dbResults {
		if _, ok := m[result.MyStatus]; !ok {
			m[result.MyStatus] = make([]dbResult, 0)
			mTotalCount[result.MyStatus] = 0
		}
		m[result.MyStatus] = append(m[result.MyStatus], result)
		mTotalCount[result.MyStatus] += result.Count
	}

	var results = make([]resps.EnvStatResp, 0)
	for k, v := range m {
		data := resps.EnvStatResp{
			Status:  k,
			Count:   mTotalCount[k],
			Details: make([]resps.DetailStatResp, 0),
		}

		for _, e := range v {
			data.Details = append(data.Details, resps.DetailStatResp{
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
func GetProjectResStat(tx *db.Session, projectId models.Id, limit int) ([]resps.ResStatResp, e.Error) {
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

	var results []resps.ResStatResp
	for k, v := range m {
		data := resps.ResStatResp{
			ResType: k,
			Count:   mTotalCount[k],
			Details: make([]resps.DetailStatResp, 0),
		}

		for _, e := range v {
			data.Details = append(data.Details, resps.DetailStatResp{
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
func GetProjectEnvResStat(tx *db.Session, projectId models.Id, limit int) ([]resps.ProjOrEnvResStatResp, e.Error) {
	/* sample sql:
	select
		iac_resource.env_id as id,
		iac_env.name as name,
		iac_resource.type as res_type,
		DATE_FORMAT(iac_resource.applied_at, "%Y-%m") as date,
		count(*) as count
	from
		iac_resource
	JOIN iac_env ON
		iac_env.last_res_task_id = iac_resource.task_id
		and iac_env.id = iac_resource.env_id
	where
		iac_env.project_id = 'p-c8gg9josm56injdlb86g'
		AND (DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(CURDATE(), "%Y-%m")
			OR
		DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), "%Y-%m"))
	group by
		date,
		iac_resource.type,
		iac_resource.env_id
	limit 10;
	*/

	query := tx.Model(&models.Resource{}).Select(`iac_resource.env_id as id, iac_env.name as name, iac_resource.type as res_type, DATE_FORMAT(iac_resource.applied_at, "%Y-%m") as date, count(*) as count`)

	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id and iac_env.id = iac_resource.env_id`)
	query = query.Where(`iac_env.project_id = ?`, projectId)
	query = query.Where(`DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(CURDATE(), "%Y-%m") OR DATE_FORMAT(applied_at, "%Y-%m") = DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), "%Y-%m")`)

	query = query.Group("date, iac_resource.type,iac_resource.env_id")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var dbResults []ProjectStatResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return dbResult2ProjectResStatResp(dbResults), nil
}

// GetProjectResGrowTrend 最近7天资源及费用趋势
func GetProjectResGrowTrend(tx *db.Session, projectId models.Id, days int) ([][]resps.ResGrowTrendResp, e.Error) {
	/* sample sql
	select
		iac_resource.env_id as id,
		iac_env.name as name,
		iac_resource.type as res_type,
		DATE_FORMAT(iac_resource.applied_at, "%Y-%m-%d") as date,
		count(*) as count
	from
		iac_resource
	JOIN iac_env ON
		iac_env.last_res_task_id = iac_resource.task_id
		and iac_env.id = iac_resource.env_id
	where
		iac_env.project_id = 'p-c8gg9josm56injdlb86g'
		and (
		DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 7 DAY), "%Y-%m-%d")
			or (DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(DATE_SUB(CURDATE(), INTERVAL 7 DAY), INTERVAL 1 MONTH), "%Y-%m-%d")
				and DATE_FORMAT(applied_at, "%Y-%m-%d") <= DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), "%Y-%m-%d")))
	group by
		date,
		iac_resource.type,
		iac_resource.env_id
	order by
		date
	*/

	query := tx.Model(&models.Resource{}).Select(`iac_resource.env_id as id, iac_env.name as name, iac_resource.type as res_type, DATE_FORMAT(iac_resource.applied_at, "%Y-%m-%d") as date, count(*) as count`)
	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id and iac_env.id = iac_resource.env_id`)

	query = query.Where("iac_env.project_id = ?", projectId)
	query = query.Where(`DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL ? DAY), "%Y-%m-%d") or (DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(DATE_SUB(CURDATE(), INTERVAL ? DAY), INTERVAL 1 MONTH), "%Y-%m-%d") and DATE_FORMAT(applied_at, "%Y-%m-%d") <= DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), "%Y-%m-%d"))`, days, days)

	query = query.Group("date, iac_resource.type, iac_resource.env_id").Order("date")

	var dbResults []ProjectStatResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	now := time.Now()
	var results = make([][]resps.ResGrowTrendResp, 2)

	startDate := now.AddDate(0, -1, -1*days)
	endDate := now.AddDate(0, -1, 0)
	var mPreDateCount map[string]int
	var mPreResTypeCount map[[2]string]int
	var mPreEnvCount map[[3]string]int
	results[0], mPreDateCount, mPreResTypeCount, mPreEnvCount = getResGrowTrendByDays(startDate, endDate, dbResults, days)

	startDate = now.AddDate(0, 0, -1*days)
	endDate = now
	var mDateCount map[string]int
	var mResTypeCount map[[2]string]int
	var mEnvCount map[[3]string]int
	results[1], mDateCount, mResTypeCount, mEnvCount = getResGrowTrendByDays(startDate, endDate, dbResults, days)

	// 计算增长量
	for i := range results[1] {
		// 每天增长量
		curDate := results[1][i].Date
		preDate := calcPreDayKey(curDate, days)
		results[1][i].Up = mDateCount[results[1][i].Date]
		if _, ok := mPreDateCount[preDate]; ok {
			results[1][i].Up -= mPreDateCount[preDate]
		}

		// 每天每个资源类型增长量
		for j := range results[1][i].ResTypes {
			resType := results[1][i].ResTypes[j].ResType
			curResKey := [2]string{curDate, resType}
			preResKey := [2]string{preDate, resType}
			results[1][i].ResTypes[j].Up = mResTypeCount[curResKey]
			if _, ok := mPreResTypeCount[preResKey]; ok {
				results[1][i].ResTypes[j].Up -= mPreResTypeCount[preResKey]
			}

			// 每天每个资源类型下每个环境增长量
			for k := range results[1][i].ResTypes[j].Details {
				envId := results[1][i].ResTypes[j].Details[k].Id.String()
				curEnvKey := [3]string{curDate, resType, envId}
				preEnvKey := [3]string{preDate, resType, envId}
				results[1][i].ResTypes[j].Details[k].Up = mEnvCount[curEnvKey]
				if _, ok := mPreResTypeCount[preResKey]; ok {
					results[1][i].ResTypes[j].Details[k].Up -= mPreEnvCount[preEnvKey]
				}
			}
		}
	}

	return results, nil
}
