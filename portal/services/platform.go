// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/resps"
)

func GetBaseDataCount(dbSess *db.Session) (*resps.PfBasedataResp, e.Error) {
	var err error
	var result = &resps.PfBasedataResp{}

	// organization
	result.OrgCount.Total, result.OrgCount.Active, err = getTotalAndActiveCount(dbSess, models.Organization{}.TableName(), models.Enable)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	// project
	result.ProjectCount.Total, result.ProjectCount.Active, err = getTotalAndActiveCount(dbSess, models.Project{}.TableName(), models.Enable)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	// enviroment
	result.EnvCount.Total, result.EnvCount.Active, err = getTotalAndActiveCount(dbSess, models.Env{}.TableName(), models.EnvStatusActive)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	// stack
	result.StackCount.Total, result.StackCount.Active, err = getTotalAndActiveCount(dbSess, models.Template{}.TableName(), models.Enable)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	// user
	result.UserCount.Total, result.UserCount.Active, err = getTotalAndActiveCount(dbSess, models.User{}.TableName(), models.Enable)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return result, nil
}

func getTotalAndActiveCount(dbSess *db.Session, tableName, status string) (int64, int64, error) {
	cntTotal, err := dbSess.Table(tableName).Count()
	if err != nil {
		return 0, 0, err
	}

	cntActive, err := dbSess.Table(tableName).Where("status = ?", status).Count()
	if err != nil {
		return 0, 0, err
	}

	return cntTotal, cntActive, nil
}

func GetProviderEnvCount(dbSess *db.Session) ([]resps.PfProEnvStatResp, e.Error) {
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
			iac_resource.env_id = iac_env.id
		where
			iac_env.archived = 0
		group by
			provider,
			env_id
	) as t
	group by
		t.provider
	*/
	subQuery := dbSess.Model(&models.Resource{}).Select(`provider, env_id`)
	subQuery = subQuery.Joins(`join iac_env ON iac_resource.env_id = iac_env.id`)
	subQuery = subQuery.Where("iac_env.archived = ?", 0)
	subQuery = subQuery.Group("provider, env_id")

	query := dbSess.Table(`(?) as t`, subQuery.Expr()).Select(`t.provider as provider, COUNT(*) as count`)
	query = query.Group("t.provider")

	var dbResults []resps.PfProEnvStatResp
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return dbResults, nil
}

func GetProviderResCount(dbSess *db.Session) ([]resps.PfProResStatResp, e.Error) {
	return nil, nil
}
