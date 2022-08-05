// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/resps"
)

func GetTodayCreatedOrgs(dbSess *db.Session) (int64, error) {
	query := dbSess.Model(&models.Organization{}).Where("status = ?", models.Enable)
	query = query.Where(`DATE_FORMAT(created_at, "%Y-%m-%d") = CURDATE()`)

	return query.Count()
}

func GetTodayCreatedProjects(dbSess *db.Session, orgIds []string) (int64, error) {
	query := dbSess.Model(&models.Project{}).Where("status = ?", models.Enable)
	query = query.Where(`DATE_FORMAT(created_at, "%Y-%m-%d") = CURDATE()`)
	if len(orgIds) > 0 {
		query = query.Where(`org_id IN (?)`, orgIds)
	}

	return query.Count()
}

func GetTodayCreatedStacks(dbSess *db.Session, orgIds []string) (int64, error) {
	query := dbSess.Model(&models.Template{}).Where("status = ?", models.Enable)
	query = query.Where(`DATE_FORMAT(created_at, "%Y-%m-%d") = CURDATE()`)
	if len(orgIds) > 0 {
		query = query.Where(`org_id IN (?)`, orgIds)
	}

	return query.Count()
}

func GetTodayCreatedPGs(dbSess *db.Session, orgIds []string) (int64, error) {
	query := dbSess.Model(&models.PolicyGroup{}).Where(`DATE_FORMAT(created_at, "%Y-%m-%d") = CURDATE()`)
	if len(orgIds) > 0 {
		query = query.Where(`org_id IN (?)`, orgIds)
	}

	return query.Count()
}

func GetTodayCreatedEnvs(dbSess *db.Session, orgIds []string) (int64, error) {
	query := dbSess.Model(&models.Env{}).Where(`DATE_FORMAT(created_at, "%Y-%m-%d") = CURDATE()`)
	if len(orgIds) > 0 {
		query = query.Where(`org_id IN (?)`, orgIds)
	}

	return query.Count()
}

func GetTodayDestroyedEnvs(dbSess *db.Session, orgIds []string) (int64, error) {
	query := dbSess.Model(&models.Env{}).Where("status = ?", models.EnvStatusDestroyed)
	query = query.Where(`DATE_FORMAT(updated_at, "%Y-%m-%d") = CURDATE()`)
	if len(orgIds) > 0 {
		query = query.Where(`org_id IN (?)`, orgIds)
	}

	return query.Count()
}

func GetTodayCreatedResTypes(dbSess *db.Session, orgIds []string) ([]resps.PfTodayResTypeStatResp, e.Error) {
	query := dbSess.Model(&models.Resource{}).Select("iac_resource.`type` as res_type, COUNT(*) as count")

	query = query.Joins(`join iac_env on iac_env.last_res_task_id = iac_resource.task_id and iac_env.id = iac_resource.env_id`)

	query = query.Where(`DATE_FORMAT(updated_at, "%Y-%m-%d") = CURDATE()`)
	if len(orgIds) > 0 {
		query = query.Where(`iac_resource.org_id IN (?)`, orgIds)
	}

	query = query.Group("iac_resource.`type`")

	var dbResults []resps.PfTodayResTypeStatResp
	if err := query.Find(&dbResults); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return dbResults, nil
}
