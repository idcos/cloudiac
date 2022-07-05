// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/resps"
	"time"
)

// activeDays 判断活跃度的天数
const activeDays = 7

func buildActiveEnvQuery(query *db.Session, selStr string, orgIds []string) *db.Session {
	query = query.Model(&models.Env{}).Select(selStr)
	query = query.Where("archived = ?", 0)
	query = query.Where(`(status = 'active' OR status = 'failed' OR task_status = 'approving' OR task_status = 'running')`)
	query = query.Where(`updated_at > DATE_SUB(CURDATE(), INTERVAL ? DAY)`, activeDays)

	return query
}

func GetOrgTotalAndActiveCount(dbSess *db.Session, orgIds []string) (int64, int64, error) {
	queryTotal := dbSess.Model(&models.Organization{}).Where(`status = ?`, models.Enable)
	if len(orgIds) > 0 {
		queryTotal = queryTotal.Where("id IN (?)", orgIds)
	}
	cntTotal, err := queryTotal.Count()
	if err != nil {
		return 0, 0, err
	}

	queryActive := dbSess.Model(&models.Organization{}).Where(`status = ?`, models.Enable)
	subQuery := buildActiveEnvQuery(dbSess, "DISTINCT(org_id)", orgIds)

	cntActive, err := queryActive.Where(`id IN (?)`, subQuery.Expr()).Count()
	if err != nil {
		return 0, 0, err
	}

	return cntTotal, cntActive, nil
}

func GetProjectTotalAndActiveCount(dbSess *db.Session, orgIds []string) (int64, int64, error) {
	queryTotal := dbSess.Model(&models.Project{}).Where(`status = ?`, models.Enable)
	if len(orgIds) > 0 {
		queryTotal = queryTotal.Where("id IN (?)", orgIds)
	}
	cntTotal, err := queryTotal.Count()
	if err != nil {
		return 0, 0, err
	}

	queryActive := dbSess.Model(&models.Project{}).Where(`status = ?`, models.Enable)
	subQuery := buildActiveEnvQuery(dbSess, "DISTINCT(project_id)", orgIds)

	cntActive, err := queryActive.Where(`id IN (?)`, subQuery.Expr()).Count()
	if err != nil {
		return 0, 0, err
	}

	return cntTotal, cntActive, nil
}

func GetEnvTotalAndActiveCount(dbSess *db.Session, orgIds []string) (int64, int64, error) {
	queryTotal := dbSess.Model(&models.Env{}).Where(`archived = ?`, 0)
	if len(orgIds) > 0 {
		queryTotal = queryTotal.Where("id IN (?)", orgIds)
	}
	cntTotal, err := queryTotal.Count()
	if err != nil {
		return 0, 0, err
	}

	queryActive := buildActiveEnvQuery(dbSess, "COUNT(*)", orgIds)

	cntActive, err := queryActive.Count()
	if err != nil {
		return 0, 0, err
	}

	return cntTotal, cntActive, nil
}

func GetStackTotalAndActiveCount(dbSess *db.Session, orgIds []string) (int64, int64, error) {
	queryTotal := dbSess.Model(&models.Template{}).Where(`status = ?`, models.Enable)
	if len(orgIds) > 0 {
		queryTotal = queryTotal.Where("id IN (?)", orgIds)
	}
	cntTotal, err := queryTotal.Count()
	if err != nil {
		return 0, 0, err
	}

	queryActive := dbSess.Model(&models.Template{}).Where(`status = ?`, models.Enable)
	subQuery := buildActiveEnvQuery(dbSess, "DISTINCT(tpl_id)", orgIds)

	cntActive, err := queryActive.Where(`id IN (?)`, subQuery.Expr()).Count()
	if err != nil {
		return 0, 0, err
	}

	return cntTotal, cntActive, nil
}

func GetUserTotalAndActiveCount(dbSess *db.Session, orgIds []string) (int64, int64, error) {
	queryTotal := dbSess.Model(&models.User{})
	queryTotal = queryTotal.Joins(`join iac_user_org on iac_user.id = iac_user_org.user_id`)
	queryTotal = queryTotal.Where(`iac_user.status = ?`, models.Enable)
	queryTotal = queryTotal.Where(`iac_user.active_status = ?`, "active")
	if len(orgIds) > 0 {
		queryTotal = queryTotal.Where("iac_user_org.org_id IN (?)", orgIds)
	}
	cntTotal, err := queryTotal.Count()
	if err != nil {
		return 0, 0, err
	}

	queryActive := queryTotal.Where(`iac_user.updated_at > DATE_SUB(CURDATE(), INTERVAL ? DAY)`, activeDays)

	cntActive, err := queryActive.Count()
	if err != nil {
		return 0, 0, err
	}

	return cntTotal, cntActive, nil
}

func GetProviderEnvCount(dbSess *db.Session, orgIds []string) ([]resps.PfProEnvStatResp, e.Error) {
	/* sample sql
	SELECT
		t.provider as provider,
		COUNT(*) as count
	FROM
			(
		select
			SUBSTRING_INDEX(provider,'/',-1) as provider,
			env_id
		from
			iac_resource
		join iac_env ON
			iac_env.last_res_task_id = iac_resource.task_id
			and iac_env.id = iac_resource.env_id
		where
			iac_env.archived = 0
		AND iac_resource.org_id IN ('xxx', 'yyy')
		group by
			provider,
			env_id
		) as t
	group by
		t.provider
	*/
	subQuery := dbSess.Model(&models.Resource{}).Select(`SUBSTRING_INDEX(provider,'/',-1) as provider, env_id`)
	subQuery = subQuery.Joins(`join iac_env ON iac_env.last_res_task_id = iac_resource.task_id AND iac_resource.env_id = iac_env.id`)
	subQuery = subQuery.Where("iac_env.archived = ?", 0)
	if len(orgIds) > 0 {
		subQuery = subQuery.Where(`iac_resource.org_id IN (?)`, orgIds)
	}
	subQuery = subQuery.Group("provider, env_id")

	query := dbSess.Table(`(?) as t`, subQuery.Expr()).Select(`t.provider as provider, COUNT(*) as count`)
	query = query.Group("t.provider")

	var dbResults []resps.PfProEnvStatResp
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return dbResults, nil
}

func GetProviderResCount(dbSess *db.Session, orgIds []string) ([]resps.PfProResStatResp, e.Error) {
	/* sample sql
	SELECT
		SUBSTRING_INDEX(provider,'/',-1) as provider,
		COUNT(*) as count
	from
		iac_resource
	join iac_env on
		iac_env.last_res_task_id = iac_resource.task_id
		and iac_env.id = iac_resource.env_id
	where iac_resource.org_id IN ('xxx', 'yyy')
	GROUP BY
		provider
	*/
	query := dbSess.Model(&models.Resource{}).Select(`SUBSTRING_INDEX(provider,'/',-1) as provider, COUNT(*) as count`)
	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id and iac_env.id = iac_resource.env_id`)
	if len(orgIds) > 0 {
		query = query.Where(`iac_resource.org_id IN (?)`, orgIds)
	}
	query = query.Group("provider")

	var dbResults []resps.PfProResStatResp
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return dbResults, nil
}

func GetResTypeCount(dbSess *db.Session, orgIds []string) ([]resps.PfResTypeStatResp, e.Error) {
	/* sample sql
	SELECT
		iac_resource.`type` as res_type,
		COUNT(*) as count
	from
		iac_resource
	join iac_env on
		iac_env.last_res_task_id = iac_resource.task_id
		and iac_env.id = iac_resource.env_id
	where iac_resource.org_id IN ('xxx', 'yyy')
	GROUP BY
		iac_resource.`type`
	*/
	query := dbSess.Model(&models.Resource{}).Select("iac_resource.`type` as res_type, COUNT(*) as count")
	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id and iac_env.id = iac_resource.env_id`)
	if len(orgIds) > 0 {
		query = query.Where(`iac_resource.org_id IN (?)`, orgIds)
	}
	query = query.Group("iac_resource.`type`")

	var dbResults []resps.PfResTypeStatResp
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return dbResults, nil
}

type orgActiveResType struct {
	OrgName string
	ResType string
	Count   int64
}

func GetOrgActiveResTypeCount(dbSess *db.Session, orgIds []string) ([]resps.PfActiveResStatResp, e.Error) {
	/* sample sql
	SELECT
		iac_org.name as org_name,
		iac_resource.`type` as res_type,
		COUNT(*) as count
	from
		iac_resource
	join iac_env on
		iac_env.last_res_task_id = iac_resource.task_id
		and iac_env.id = iac_resource.env_id
	join iac_org  on
		iac_org.id = iac_resource.org_id
	where iac_resource.org_id IN ('xxx', 'yyy')
	GROUP BY
		iac_org.id,
		iac_resource.`type`
	*/

	query := dbSess.Model(&models.Resource{}).Select("iac_org.name as org_name, iac_resource.`type` as res_type, COUNT(*) as count")
	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id and iac_env.id = iac_resource.env_id`)
	query = query.Joins(`join iac_org on iac_org.id = iac_resource.org_id`)
	if len(orgIds) > 0 {
		query = query.Where(`iac_resource.org_id IN (?)`, orgIds)
	}
	query = query.Group("iac_org.id, iac_resource.`type`")

	var dbResults []orgActiveResType
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return convertToPfActiveResStatResp(dbResults), nil
}

func convertToPfActiveResStatResp(dbResults []orgActiveResType) []resps.PfActiveResStatResp {
	mOrgResTypeStat := make(map[string][]resps.PfResTypeStatResp)
	for _, dbResult := range dbResults {
		if _, ok := mOrgResTypeStat[dbResult.OrgName]; !ok {
			mOrgResTypeStat[dbResult.OrgName] = make([]resps.PfResTypeStatResp, 0)
		}
		mOrgResTypeStat[dbResult.OrgName] = append(mOrgResTypeStat[dbResult.OrgName], resps.PfResTypeStatResp{
			ResType: dbResult.ResType,
			Count:   dbResult.Count,
		})
	}

	results := make([]resps.PfActiveResStatResp, 0)
	for k, v := range mOrgResTypeStat {
		results = append(results, resps.PfActiveResStatResp{
			OrgName:      k,
			ResTypesStat: v,
		})
	}
	return results
}

func GetResWeekChange(dbSess *db.Session, orgIds []string) ([]resps.PfResWeekChangeResp, e.Error) {
	/* sample sql
	select
		DATE_FORMAT(iac_resource.applied_at, "%Y-%m-%d") as date,
		count(DISTINCT iac_resource.env_id, iac_resource.address) as count
	from
		iac_resource
	JOIN iac_env ON
		iac_env.id = iac_resource.env_id
	where
		iac_resource.org_id IN ('xxx', 'yyy')
		and DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 7 DAY), "%Y-%m-%d")
	group by
		date
	order by
		date
	*/
	query := dbSess.Model(&models.Resource{}).Select(`DATE_FORMAT(iac_resource.applied_at, "%Y-%m-%d") as date, count(DISTINCT iac_resource.env_id, iac_resource.address) as count`)
	query = query.Joins(`JOIN iac_env ON iac_env.id = iac_resource.env_id`)
	query = query.Where(`DATE_FORMAT(applied_at, "%Y-%m-%d") > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL ? DAY), "%Y-%m-%d")`, activeDays)
	if len(orgIds) > 0 {
		query = query.Where(`iac_resource.org_id IN (?)`, orgIds)
	}
	query = query.Group("date")
	query = query.Order("date")

	var dbResults []resps.PfResWeekChangeResp
	if err := query.Debug().Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return completeWeekChange(dbResults), nil
}

func completeWeekChange(dbResults []resps.PfResWeekChangeResp) []resps.PfResWeekChangeResp {
	var mWeekChange = make(map[string]int64)
	for _, result := range dbResults {
		mWeekChange[result.Date] = result.Count
	}

	fullResults := make([]resps.PfResWeekChangeResp, 0)
	startDate := time.Now().AddDate(0, 0, -1*activeDays)
	for i := 0; i < activeDays; i++ {
		startDate = startDate.AddDate(0, 0, 1)
		date := startDate.Format("2006-01-02")
		var count int64 = 0

		if v, ok := mWeekChange[date]; ok {
			count = v
		}

		fullResults = append(fullResults, resps.PfResWeekChangeResp{
			Date:  date,
			Count: count,
		})
	}

	return fullResults

}
