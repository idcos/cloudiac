// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

type ResourceDrift struct {
	TimedModel
	ResId       Id     `json:"resId" gorm:"size:32;not null"`
	DriftDetail string `json:"driftDetail" gorm:"type:text"`
}

func (ResourceDrift) TableName() string {
	return "iac_resource_drift"
}
