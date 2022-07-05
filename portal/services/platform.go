// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/resps"
	"strings"
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
			provider,
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
	subQuery := dbSess.Model(&models.Resource{}).Select(`provider, env_id`)
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

	// provider 名称截取最后一段
	for i, result := range dbResults {
		providers := strings.Split(result.Provider, "/")
		dbResults[i].Provider = providers[len(providers)-1]
	}

	return dbResults, nil
}

func GetProviderResCount(dbSess *db.Session) ([]resps.PfProResStatResp, e.Error) {
	return nil, nil
}
