// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"fmt"
)

func CreateTemplateProject(tx *db.Session, projectIds []models.Id, tplId models.Id) e.Error {
	bq := utils.NewBatchSQL(1024, "INSERT INTO", models.ProjectTemplate{}.TableName(),
		"template_id", "project_id")

	for _, v := range projectIds {
		if err := bq.AddRow(tplId, v); err != nil {
			return e.New(e.DBError, err)
		}
	}

	for bq.HasNext() {
		sql, args := bq.Next()
		if _, err := tx.Exec(sql, args...); err != nil {
			return e.New(e.DBError, err)
		}
	}
	return nil
}

func DeleteTemplateProject(tx *db.Session, id models.Id) e.Error {
	if _, err := tx.Where("template_id = ?", id).Delete(&models.ProjectTemplate{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete project_template error: %v", err))
	}
	return nil
}
