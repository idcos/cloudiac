// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"net/http"
)

func DeletePolicyGroupRel(tx *db.Session, id models.Id, scope string) e.Error {
	sql := ""
	if scope == consts.ScopeEnv {
		sql = "env_id = ? and group_id != ''"
	} else {
		sql = "tpl_id = ? and env_id = '' and group_id != ''"
	}
	if _, err := tx.Where(sql, id).Delete(models.PolicyRel{}); err != nil {
		if e.IsRecordNotFound(err) {
			return nil
		}
		return e.New(e.DBError, err)
	}
	return nil
}

func GetPolicyRel(query *db.Session, id models.Id, scope string) (*models.PolicyRel, e.Error) {
	sql := ""
	if scope == consts.ScopeEnv {
		sql = "env_id = ? and group_id = ''"
	} else {
		sql = "tpl_id = ? and group_id = ''"
	}
	rel := models.PolicyRel{}
	if err := query.Model(models.PolicyRel{}).Where(sql, id).First(&rel); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.PolicyRelNotExist, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &rel, nil
}

type PolicyGroupsNameResp struct {
	models.PolicyRel
	PolicyGroupId string `json:"policyGroupId"`
}

func GetPolicyRels(db *db.Session, id models.Id, scope string) ([]*PolicyGroupsNameResp, e.Error) {
	sql := ""
	if scope == consts.ScopeEnv {
		sql = "env_id = ?"
	} else {
		sql = "tpl_id = ?"
	}
	rel := []*PolicyGroupsNameResp{}
	query := db.Model(rel).Joins("left join iac_policy_group on iac_policy_rel.group_id = iac_policy_group.id").
		LazySelectAppend("iac_policy_group.id as policy_group_id, iac_policy_rel.*")

	if err := query.Where(sql, id).Scan(&rel); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.PolicyRelNotExist, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return rel, nil
}

func CreatePolicyRel(tx *db.Session, rel *models.PolicyRel) (*models.PolicyRel, e.Error) {
	if err := models.Create(tx, rel); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.PolicyRelAlreadyExist, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return rel, nil
}

func DeletePolicyEnabledRel(tx *db.Session, id models.Id, scope string) e.Error {
	sql := ""
	if scope == consts.ScopeEnv {
		sql = "env_id = ? and group_id = ''"
	} else {
		sql = "tpl_id = ? and env_id = '' and group_id = ''"
	}
	if _, err := tx.Where(sql, id).Delete(models.PolicyRel{}); err != nil {
		if e.IsRecordNotFound(err) {
			return nil
		}
		return e.New(e.DBError, err)
	}
	return nil
}

// UpdatePolicyRel 创建/更新策略关系
func UpdatePolicyRel(tx *db.Session, form *forms.UpdatePolicyRelForm) ([]*models.PolicyRel, e.Error) {
	var (
		env  *models.Env
		tpl  *models.Template
		rels []*models.PolicyRel
		err  e.Error
	)
	if form.Scope == consts.ScopeEnv {
		env, err = GetEnvById(tx, form.Id)
		if err != nil {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
	} else {
		tpl, err = GetTemplateById(tx, form.Id)
		if err != nil {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
	}

	// 删除原有关联关系
	if err := DeletePolicyGroupRel(tx, form.Id, form.Scope); err != nil && !e.IsRecordNotFound(err) {
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 创新新的关联关系
	for _, groupId := range form.PolicyGroupIds {
		group, err := GetPolicyGroupById(tx, groupId)
		if err != nil {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
		if env != nil {
			rels = append(rels, &models.PolicyRel{
				OrgId:     env.OrgId,
				ProjectId: env.ProjectId,
				GroupId:   group.Id,
				EnvId:     env.Id,
				Scope:     consts.ScopeEnv,
			})
		} else {
			rels = append(rels, &models.PolicyRel{
				OrgId:   tpl.OrgId,
				GroupId: group.Id,
				TplId:   tpl.Id,
				Scope:   models.PolicyRelScopeTpl,
			})
		}
	}

	if len(rels) > 0 {
		if er := models.CreateBatch(tx, rels); er != nil {
			return nil, e.New(e.DBError, er)
		}
	}
	return rels, nil
}

func DeleteRelByPolicyGroupId(tx *db.Session, groupId models.Id) e.Error {
	if _, err := tx.Where("group_id = ?", groupId).Delete(models.PolicyRel{}); err != nil {
		if e.IsRecordNotFound(err) {
			return nil
		}
		return e.New(e.DBError, err)
	}
	return nil
}
