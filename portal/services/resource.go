// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func GetResourceByEnvId(tx *db.Session, envId models.Id) (models.ResFields, error) {
	resources := models.ResFields{}
	if err := tx.Raw("SELECT res_id,any_value(applied_at) as applied_at FROM (SELECT * FROM iac_resource ORDER BY applied_at LIMIT 9999999) res Where env_id = ? GROUP BY res_id;", envId).Scan(&resources); err != nil {
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
