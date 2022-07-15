// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func GetTodayCreatedOrgs(dbSess *db.Session, orgIds []string) (int64, error) {
	query := dbSess.Model(&models.Organization{}).Where("status = ?", models.Enable)
	query = query.Where(`DATE_FORMAT(created_at, "%Y-%m-%d") = CURDATE()`)

	return query.Count()
}
