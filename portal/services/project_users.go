package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/utils"
)

func BindProjectUsers(tx *db.Session, projectId models.Id, authorization []forms.UserAuthorization) e.Error {
	bq := utils.NewBatchSQL(1024, "INSERT INTO", models.UserProject{}.TableName(),
		"user_id", "project_id", "role")

	for _, v := range authorization {
		if err := bq.AddRow(v.UserId, projectId, v.Role); err != nil {
			return e.New(e.DBError, err)
		}
	}

	for bq.HasNext() {
		sql, args := bq.Next()
		if _, err := tx.Exec(sql, args); err != nil {
			return e.New(e.DBError, err)
		}
	}
	return nil
}

func UpdateProjectUsers(tx *db.Session, projectId models.Id, authorization []forms.UserAuthorization) e.Error {
	if _, err := tx.Where("project_id = ?", projectId).Delete(&models.UserProject{}); err != nil {
		return e.New(e.DBError, err)
	}
	return BindProjectUsers(tx, projectId, authorization)
}
