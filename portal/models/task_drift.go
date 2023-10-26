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
	EnvId    Id     `json:"envId" gorm:"index;size:32;not null"`
	TaskId   Id     `json:"taskId" gorm:"index;size:32;not null"`
	Type     string `json:"type" gorm:"not null"`
	IsDrift  bool   `json:"isDrift" gorm:"default:false"`
	ExecTime Time   `json:"execTime" gorm:"type:datetime" example:"2006-01-02 15:04:05"` // 执行时间
}

func (TaskDrift) TableName() string {
	return "iac_task_drift"
}

type EnvDrift struct {
	EnvId             Id     `json:"envId"`             // 环境id
	IsDrift           bool   `json:"isDrift"`           // 是否偏移
	CronDriftExpress  string `json:"cronDriftExpress"`  // 偏移检测任务的Cron表达式
	AutoRepairDrift   bool   `json:"autoRepairDrift"`   // 是否进行自动纠偏
	OpenCronDrift     bool   `json:"openCronDrift"`     // 是否开启偏移检测
	DriftTime         *Time  `json:"driftTime"`         // 最后一次检测时间
	DriftStatus       string `json:"driftStatus"`       // 最后一次检测状态
	NextDriftTaskTime *Time  `json:"nextDriftTaskTime"` // 下次检测时间
}

type TaskDriftInfo struct {
	TaskDrift
	Status string `json:"status"` // 漂移任务结果
}
