// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

func GetResourceByEnvId(tx *db.Session, envId models.Id) (models.ResFields, error) {
	resources := models.ResFields{}
	if err := tx.Raw("select res_id,min(applied_at) as applied_at from iac_resource where env_id = ? and applied_at is not null group by res_id;", envId).Scan(&resources); err != nil {
		return nil, err
	}
	if len(resources) == 0 {
		return nil, nil
	} else {
		return resources, nil
	}
}

func SetResFieldsAsMap(field models.ResFields) map[string]interface{} {
	if field == nil {
		return nil
	}
	resources := make(map[string]interface{})
	for _, res := range field {
		resources[string(res.ResId)] = res.AppliedAt
	}
	return resources
}

func GetResourceByIdsInProvider(dbSess *db.Session, ids, projectIds []string, vg models.VariableGroup) ([]models.Resource, e.Error) {
	resp := make([]models.Resource, 0)
	query := dbSess.Model(models.Resource{}).
		Where("provider like ?", fmt.Sprintf("%%%s", vg.Provider)).
		Where("org_id = ?", vg.OrgId).
		Where("res_id in (?)", ids).
		Group("res_id").Group("env_id").
		Select("res_id")

	if !(len(projectIds) == 1 && projectIds[0] == "") {
		query = query.Where("project_id  in  (?)", projectIds)
	}

	if err := dbSess.Model(models.Resource{}).
		Where("res_id in (?)", query.Expr()).
		Find(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return resp, nil
}
