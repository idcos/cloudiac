
package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

func CreateTemplate(tx *db.Session, template models.Template) (*models.Template, e.Error) {
	if template.Id == "" {
		template.Id = models.NewId("ct")
	}
	if err := models.Create(tx, &template); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.TemplateAlreadyExists, err)
	}
		return nil, e.New(e.DBError, err)
	}
	return &template, nil
}

func UpdateTemplate(tx *db.Session, id models.Id, attrs models.Attrs) (tpl *models.Template, re e.Error) {
	tpl = &models.Template{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Template{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.UserEmailDuplicate)
		}
		return nil, e.New(e.DBError, fmt.Errorf("update template error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(tpl); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query template error: %v", err))
	}
	return
}

func DeleteTemplate(tx *db.Session, id models.Id) e.Error {
	if _, err := tx.Where("id = ?", id).Delete(&models.Template{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete template error: %v", err))
	}
	return nil
}

func GetTemplateById(tx *db.Session, id models.Id) (*models.Template, e.Error) {
	tpl := models.Template{}
	if err := tx.Where("id = ?", id).First(&tpl); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TemplateNotExists, err)
		}
	}
	return &tpl, nil

}