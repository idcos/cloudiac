// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
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
