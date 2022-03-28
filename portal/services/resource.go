// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func getResourceByEnvId(tx *db.Session, envId models.Id) ([]models.Resource, error) {
	resources := make([]models.Resource, 0)
	if err := tx.Table(models.Resource{}.TableName()).Where("env_id", envId).Order("applied_at").Find(&resources); err != nil {
		return nil, err
	}
	if len(resources) == 0 {
		return nil, nil
	} else {
		return resources, nil
	}
}
