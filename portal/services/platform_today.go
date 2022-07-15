// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
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
