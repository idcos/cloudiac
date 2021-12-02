// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
	"strings"
)

func CreateTemplate(tx *db.Session, tpl models.Template) (*models.Template, e.Error) {
	if tpl.Id == "" {
		tpl.Id = tpl.NewId()
	}

	if strings.HasPrefix(tpl.Workdir, "..") || strings.HasPrefix(tpl.Workdir, "/") {
		return nil, e.New(e.BadParam, fmt.Errorf("invalid workdir '%s'", tpl.Workdir))
	}

	if err := models.Create(tx, &tpl); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.TemplateAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &tpl, nil
}

func UpdateTemplate(tx *db.Session, id models.Id, attrs models.Attrs) (tpl *models.Template, re e.Error) {
	tpl = &models.Template{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Template{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.TemplateAlreadyExists, err)
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

func QueryTemplateByOrgId(tx *db.Session, q string, orgId models.Id, templateIdList []models.Id) *db.Session {
	query := tx.Model(&models.Template{}).Joins(
		"LEFT  JOIN iac_user"+
			"  ON iac_user.id = iac_template.creator_id").
		LazySelectAppend(
			"iac_user.name as creator",
			"iac_template.*")
	query = query.Joins("left join iac_env on iac_template.id = iac_env.tpl_id and (iac_env.status != 'inactive' or deploying = 1)").Group("iac_template.id").
		LazySelectAppend("count(iac_env.id) as active_environment")
	if q != "" {
		qs := "%" + q + "%"
		query = query.Where("iac_template.name LIKE ? OR iac_template.description LIKE ?", qs, qs)
	}
	query = query.Where("iac_template.org_id = ?", orgId).Order("iac_template.created_at DESC")
	if len(templateIdList) != 0 {
		// 如果传入项目id，需要项目ID 再次筛选
		query = query.Where("iac_template.id in (?) ", templateIdList)
	}
	return query.LazySelectAppend("iac_template.tpl_type")
}

func QueryTplByProjectId(tx *db.Session, projectId models.Id) (tplIds []models.Id, err e.Error) {
	if err := tx.Table(models.ProjectTemplate{}.TableName()).
		Where("project_id = ?", projectId).
		Pluck("template_id", &tplIds); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return
}

func QueryProjectByTplId(tx *db.Session, tplId models.Id) (projectIds []models.Id, err e.Error) {
	if err := tx.Table(models.ProjectTemplate{}.TableName()).
		Where("template_id = ?", tplId).
		Pluck("project_id", &projectIds); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return
}

func QueryTemplateByVcsIdAndRepoId(tx *db.Session, vcsId, repoId string) ([]models.Template, e.Error) {
	tpl := make([]models.Template, 0)
	if err := tx.Where("vcs_id = ?", vcsId).
		Where("repo_id = ?", repoId).
		Find(&tpl); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TemplateNotExists, err)
		}
	}
	return tpl, nil
}

func QueryTplByVcsId(tx *db.Session, VcsId models.Id) (bool, e.Error) {
	exists, err := tx.Table(models.Template{}.TableName()).
		Where("vcs_id = ? and deleted_at_t = 0", VcsId).Exists()
	if err != nil {
		return false, e.AutoNew(err, e.DBError)
	}
	return exists, nil
}

func GetTplLastScanTask(sess *db.Session, tplId models.Id) (*models.ScanTask, error) {
	task := models.ScanTask{}
	scanTaskIdQuery := sess.Model(&models.Template{}).Where("id = ?", tplId).Select("last_scan_task_id")
	err := sess.Model(&models.ScanTask{}).Where("id = (?)", scanTaskIdQuery.Expr()).First(&task)
	return &task, err
}

func QueryTemplateByName(tx *db.Session, name string, OrgId models.Id) (*models.Template, e.Error) {
	tpl := models.Template{}
	if err := tx.Where("name = ? and org_id = ?", name, OrgId).First(&tpl); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TemplateNotExists, err)
		}
	}
	return &tpl, nil
}

func QueryTemplate(tx *db.Session) *db.Session {
	return tx.Model(&models.Template{})
}

// 通过名称查询指定组织下的模板
func FindOrgTemplateByName(tx *db.Session, orgId models.Id, name string) (tpl models.Template, err error) {
	err = tx.Model(&models.Template{}).Where("org_id = ? AND name = ?", orgId, name).Find(&tpl)
	return tpl, err
}

func GetTplByEnvId(sess *db.Session, envId models.Id) (*models.Template, e.Error) {
	env, err := GetEnvById(sess, envId)
	if err != nil {
		return nil, err
	}
	return GetTemplateById(sess, env.TplId)

}
