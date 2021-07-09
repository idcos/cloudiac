package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func GetTemplate(sess *db.Session, id models.Id) (*models.Template, error) {
	tpl := models.Template{}
	err := sess.Where("id = ?", id).First(&tpl)
	return &tpl, err
}

func GetTemplateById(tx *db.Session, id models.Id) (*models.Template, e.Error) {
	o := models.Template{}
	if err := tx.Where("id = ?", id).First(&o); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TemplateNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &o, nil
}
