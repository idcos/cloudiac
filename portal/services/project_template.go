package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
)

func CreateTemplateProject(tx *db.Session, projectIds []models.Id, tplId models.Id) e.Error {
	bq := utils.NewBatchSQL(1024, "INSERT INTO", models.ProjectTemplate{}.TableName(),
		"template_id", "project_id", "role")

	for _, v := range projectIds {
		if err := bq.AddRow(v, tplId); err != nil {
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
