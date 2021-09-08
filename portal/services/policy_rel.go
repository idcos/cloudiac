// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func DeletePolicyRel(tx *db.Session, id models.Id, scope string) e.Error {
	sql := ""
	if scope == consts.ScopeEnv {
		sql = "env_id = ?"
	} else {
		sql = "tpl_id = ? and env_id = ''"
	}
	if _, err := tx.Where(sql, id).Delete(models.PolicyRel{}); err != nil {
		if e.IsRecordNotFound(err) {
			return nil
		}
		return e.New(e.DBError, err)
	}
	return nil
}
