package services

import (
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"fmt"
)

func CreateTemplate(tx *db.Session, template models.Template) (*models.Template, e.Error) {
	if err := models.Create(tx, &template); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.TemplateAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &template, nil
}

func QueryTemplate(tx *db.Session, status, q, taskStatus string, statusList []string, orgId uint) (*db.Session, *db.Session) {
	query := tx.Debug().Model(&models.Template{}).Joins(
		"LEFT  JOIN iac_task" +
			"  ON iac_task.template_guid = iac_template.guid" +
			" AND iac_task.id IN " +
			"(SELECT MAX(id)  FROM iac_task  GROUP BY template_id)").
		LazySelectAppend(
			"iac_task.updated_at as task_updated_at",
			"iac_task.guid as task_guid",
			"iac_task.`status` as task_status",
			"iac_template.*")
	if taskStatus != "all" && taskStatus != "" {
		query = query.Where("iac_task.`status` in (?)", statusList)
	}
	if status != "" {
		query = query.Where("iac_template.`status` = ?", status)
	}
	if q != "" {
		qs := "%" + q + "%"
		query = query.Where("iac_template.name LIKE ? OR iac_template.description LIKE ?", qs, qs)
	}

	query = query.Where("iac_template.org_id = ?", orgId).Order("iac_template.created_at DESC")
	return query, query
}

func UpdateTemplate(tx *db.Session, id uint, attrs models.Attrs) (tpl *models.Template, re e.Error) {
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


func DetailTemplate(tx *db.Session, tId uint) (models.Template, e.Error) {
	tpl := models.Template{}
	if err := tx.Table(models.Template{}.TableName()).Where("id = ?", tId).First(&tpl); err != nil {
		return tpl, e.New(e.DBError, err)
	}
	return tpl, nil
}

func OverviewTemplate(tx *db.Session, tId uint) *db.Session {
	return tx.Table(models.Template{}.TableName()).Where("id = ?", tId)
}

func OverviewTemplateTask(tx *db.Session, tId uint) (task []models.Task, err e.Error) {
	if err := tx.Table(models.Task{}.TableName()).
		Where("template_id = ?", tId).
		//Where("start_at is not null").
		//Where("end_at is not null").
		Order("updated_at desc").
		//Limit(3).
		Find(&task); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return
}

func GetTemplateByGuid(tx *db.Session, tGuid string) (models.Template, e.Error) {
	tpl := models.Template{}
	if err := tx.Table(models.Template{}.TableName()).
		Where("guid = ?", tGuid).First(&tpl); err != nil {
		return models.Template{}, e.New(e.DBError, fmt.Errorf("template '%s' not found", tGuid))
	}
	return tpl, nil
}

func GetTemplateById(tx *db.Session, tId uint) (models.Template, e.Error) {
	tpl := models.Template{}
	if err := tx.Table(models.Template{}.TableName()).
		Where("id = ?", tId).
		//Limit(3).
		Find(&tpl); err != nil {
		return models.Template{}, e.New(e.DBError, err)
	}
	return tpl, nil
}
