// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

type ResourceDrift struct {
	TimedModel
	ResId       Id     `json:"resId" gorm:"size:32;not null"`
	DriftDetail string `json:"driftDetail" gorm:"type:text"`
	TaskId      Id     `json:"taskId" gorm:"index;size:32;not null"`
	IsLast      bool   `json:"isLast" gorm:"default:false"`
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

type EnvDrift struct {
	EnvId            Id     `json:"envId"`                                // 环境id
	IsDrift          bool   `json:"isDrift"`                              // 是否偏移
	CronDriftExpress string `json:"cronDriftExpress" gorm:"default:''"`   // 偏移检测任务的Cron表达式
	AutoRepairDrift  bool   `json:"autoRepairDrift" gorm:"default:false"` // 是否进行自动纠偏
	OpenCronDrift    bool   `json:"openCronDrift" gorm:"default:false"`   // 是否开启偏移检测
}
