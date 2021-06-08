package services

import (
	//"cloudiac/configs"
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
)

func SearchTemplateLibrary(query *db.Session) *db.Session {
	return query.Table(models.TemplateLibrary{}.TableName()).Order("created_at DESC")
}

func GetTemplateLibraryById(query *db.Session, id uint) (models.TemplateLibrary, e.Error) {
	tplLib := models.TemplateLibrary{}
	if err := query.Table(models.TemplateLibrary{}.TableName()).Where("id = ?", id).First(&tplLib); err != nil {
		return models.TemplateLibrary{}, e.New(e.DBError, err)
	}

	return tplLib, nil
}
