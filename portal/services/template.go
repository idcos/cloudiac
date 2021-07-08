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

// TODO 这里要不要和下面的 GetTemplate合并, 主要是错误处理
func GetTemplateById(tx *db.Session, id models.Id) (*models.Template, e.Error) {
	tpl := models.Template{}
	if err := tx.Where("id = ?", id).First(&tpl); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TemplateNotExists, err)
		}
	}
	return &tpl, nil

}
// 这是根据orgId 去查
func QueryTemplate(tx *db.Session, q string, orgId models.Id, templateIdList []string) (*db.Session, *db.Session) {
	query := tx.Debug().Model(&models.Template{}).Joins(
		"LEFT  JOIN iac_user"+
			"  ON iac_user.id = iac_template.creatorId").
		LazySelectAppend(
			"iac_user.name as iac_user_name",
			"iac_user.id",
			"iac_template.*")
	if q != "" {
		qs := "%" + q + "%"
		query = query.Where("iac_template.name LIKE ? OR iac_template.description LIKE ?", qs, qs)
	}
	query = query.Where("iac_template.org_id = ?", orgId).Order("iac_template.created_at DESC")
	if len(templateIdList) != 0 {
		// 如果传入项目id，需要项目ID 再次筛选
		query = query.Where("iac_template.id in (?) ", templateIdList)
	}
	return query, query
}


func QueryTplByProjectId(tx *db.Session, projectId models.Id) (result []string, err e.Error) {
	pro_tpl := []models.ProjectTemplate{}
	if err := tx.Where("projectId = ?", projectId).Find(&pro_tpl); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	for _, v := range pro_tpl{
		result = append(result, v.TemplateId)
	}
	return
}



func GetTemplate(sess *db.Session, id models.Id) (*models.Template, error) {
	tpl := models.Template{}
	err := sess.Where("id = ?", id).Find(&tpl)
	return &tpl, err
}
