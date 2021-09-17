// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
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
