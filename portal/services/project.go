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
	EnvActive    int64
	EnvFailed    int64
	EnvInactive  int64
	EnvDestroyed int64
}, error) {
	var (
		resp []struct {
			Count  int64
			Status string
		}
		envActive    int64
		envFailed    int64
		envInactive  int64
		envDestroyed int64
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
		case models.EnvStatusDestroyed:
			envDestroyed = v.Count
		}
	}

	return &struct {
		EnvActive    int64
		EnvFailed    int64
		EnvInactive  int64
		EnvDestroyed int64
	}{
		EnvActive:    envActive,
		EnvFailed:    envFailed,
		EnvInactive:  envInactive,
		EnvDestroyed: envDestroyed,
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

	query := tx.Model(&models.Env{}).Select(`if(task_status = '', status, task_status) as my_status, id, name, count(*) as count`)
	query = query.Where("archived = ?", 0).Where("project_id = ?", projectId)
	query = query.Group("my_status, id")

	var dbResults []EnvStatResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return dbResult2EnvStatResp(dbResults), nil
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

	var dbResults []ResStatResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return dbResult2ResStatResp(dbResults), nil
}

// GetProjectEnvResStat 环境资源数量
func GetProjectEnvResStat(tx *db.Session, projectId models.Id, limit int) ([]resps.ProjOrEnvResStatResp, e.Error) {
	/* sample sql:
	select
		iac_resource.env_id as id,
		iac_env.name as name,
		iac_resource.type as res_type,
		count(*) as count
	from
		iac_resource
	JOIN iac_env ON
		iac_env.last_res_task_id = iac_resource.task_id
		and iac_env.id = iac_resource.env_id
	where
		iac_env.project_id = 'p-c8gg9josm56injdlb86g'
	group by
		iac_resource.type,
		iac_resource.env_id
	limit 10;
	*/

	query := tx.Model(&models.Resource{}).Select(`iac_resource.env_id as id, iac_env.name as name, iac_resource.type as res_type, count(*) as count`)

	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id and iac_env.id = iac_resource.env_id`)
	query = query.Where(`iac_env.project_id = ?`, projectId)

	query = query.Group("iac_resource.type,iac_resource.env_id")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var dbResults []ProjectOrEnvStatResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return dbResult2ProjectOrEnvResStatResp(dbResults), nil
}

// GetProjectResGrowTrend 最近7天资源及费用趋势
func GetProjectResGrowTrend(tx *db.Session, projectId models.Id, days int) ([]resps.ResGrowTrendResp, e.Error) {
	/* sample sql
	select
		iac_resource.env_id as id,
		iac_env.name as name,
		DATE_FORMAT(iac_resource.applied_at, "%Y-%m-%d") as date,
		count(DISTINCT iac_resource.env_id, iac_resource.address) as count
	from
		iac_resource
	JOIN iac_env ON
		iac_env.id = iac_resource.env_id
	where
		iac_env.project_id = 'p-c8gg9josm56injdlb86g'
		and DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 7 DAY), "%Y-%m-%d")
	group by
		date,
		iac_resource.env_id
	order by
		date
	*/

	query := tx.Model(&models.Resource{}).Select(`iac_resource.env_id as id, iac_env.name as name, DATE_FORMAT(iac_resource.applied_at, "%Y-%m-%d") as date, count(DISTINCT iac_resource.env_id, iac_resource.address) as count`)
	query = query.Joins(`join iac_env on iac_env.id = iac_resource.env_id`)

	query = query.Where("iac_env.project_id = ?", projectId)
	query = query.Where(`DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL ? DAY), "%Y-%m-%d")`, days)

	query = query.Group("date, iac_resource.env_id").Order("date")

	var dbResults []ProjectOrEnvStatResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	now := time.Now()
	startDate := now.AddDate(0, 0, -1*days)
	endDate := now

	return getResGrowTrendByDays(startDate, endDate, dbResults, days), nil
}

// GetResGrowTrendByProjects 获取项目的资源变化趋势
func GetResGrowTrendByProjects(tx *db.Session, projectIds []models.Id, days int) (map[models.Id][]resps.ProjectResStatResp, e.Error) {
	/* sample sql
	select
		iac_resource.project_id as id,
		DATE_FORMAT(iac_resource.applied_at, "%Y-%m-%d") as date,
		count(DISTINCT iac_resource.env_id, iac_resource.address) as count
	from
		iac_resource
	JOIN iac_env ON
		and iac_env.id = iac_resource.env_id
	where
		iac_env.project_id IN ('p-c9cjgrosm56nr7049qpg')
		and DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 7 DAY), "%Y-%m-%d")
	group by
		iac_resource.project_id, date
	order by
		date
	*/

	if len(projectIds) <= 0 {
		return make(map[models.Id][]resps.ProjectResStatResp), nil
	}

	query := tx.Model(&models.Resource{}).Select(`iac_resource.project_id as id, DATE_FORMAT(iac_resource.applied_at, "%Y-%m-%d") as date, count(DISTINCT iac_resource.env_id, iac_resource.address) as count`)
	query = query.Joins(`join iac_env on iac_env.id = iac_resource.env_id`)

	query = query.Where("iac_env.project_id in ?", projectIds)
	query = query.Where(`DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL ? DAY), "%Y-%m-%d")`, days)

	query = query.Group("iac_resource.project_id, date").Order("date")

	var dbResults []ProjectOrEnvStatResult
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	// mProjDateCount: projectId -> date -> count
	var mProjDateCount = make(map[models.Id]map[string]int)
	for _, dbResult := range dbResults {
		if _, ok := mProjDateCount[dbResult.Id]; !ok {
			mProjDateCount[dbResult.Id] = make(map[string]int)
		}

		mProjDateCount[dbResult.Id][dbResult.Date] = dbResult.Count
	}

	return completeDays(mProjDateCount, days), nil
}

func completeDays(m map[models.Id]map[string]int, days int) map[models.Id][]resps.ProjectResStatResp {
	var results = make(map[models.Id][]resps.ProjectResStatResp)

	now := time.Now()
	// loop projects
	for k, v := range m {
		results[k] = make([]resps.ProjectResStatResp, 0)

		startDate := now.AddDate(0, 0, -1*days)
		// loop days
		for i := 0; i < days; i++ {
			startDate = startDate.AddDate(0, 0, 1)
			dateStr := startDate.Format("2006-01-02")

			// 日期不存在，count默认值是0
			count := 0
			if _, ok := v[dateStr]; ok {
				count = m[k][dateStr]
			}

			results[k] = append(results[k], resps.ProjectResStatResp{
				Date:  dateStr,
				Count: count,
			})
		}
	}

	return results
}

// GetProjectActiveEnvs 获取项目的活跃环境数量
func GetProjectActiveEnvs(tx *db.Session, projectIds []models.Id) (map[models.Id]int, e.Error) {
	m := make(map[models.Id]int)
	if len(projectIds) <= 0 {
		return m, nil
	}

	query := tx.Model(&models.Project{}).Select(`iac_project.id as id, COUNT(*) as count `)
	query = query.Joins(`join iac_env on iac_env.project_id = iac_project.id`)
	query = query.Where(`iac_env.status = 'active'`)
	query = query.Group("iac_project.id")

	var dbResults []struct {
		Id    string
		Count int
	}

	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	for _, result := range dbResults {
		m[models.Id(result.Id)] = result.Count
	}

	return m, nil
}
