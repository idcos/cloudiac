// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/models"
	"errors"
	"gorm.io/gorm"
)

func getResourceByEnvAndResId(db *gorm.DB, envId models.Id, resId models.Id) (*models.Resource, error) {
	resource := models.Resource{}
	if err := db.Where("env_id", envId).Where("res_id", resId).Order("applied_at").First(&resource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &resource, nil
}
