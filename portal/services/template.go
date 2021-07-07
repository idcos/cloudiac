package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func GetTemplate(sess *db.Session, id models.Id) (*models.Template, error) {
	tpl := models.Template{}
	err := sess.Where("id = ?", id).First(&tpl)
	return &tpl, err
}
