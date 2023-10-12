// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

type ResourceDrift struct {
	TimedModel
	ResId       Id     `json:"resId" gorm:"size:32;not null"`
	DriftDetail string `json:"driftDetail" gorm:"type:text"`
	TaskId      Id     `json:"taskId" gorm:"index;size:32;not null"`
}

func (ResourceDrift) TableName() string {
	return "iac_resource_drift"
}

// TaskDrift 漂移检测历史
type TaskDrift struct {
	TimedModel
	EnvId   Id     `json:"envId" gorm:"index;size:32;not null"`
	TaskId  Id     `json:"taskId" gorm:"index;size:32;not null"`
	Type    string `json:"type" gorm:"not null"`
	Status  string `json:"status" gorm:"not null"`
	IsDrift bool   `json:"isDrift" gorm:"default:false"`
}

func (TaskDrift) TableName() string {
	return "iac_task_drift"
}
