package models

import "cloudiac/portal/libs/db"

type ResourceDrift struct {
	AutoUintIdModel
	EnvId          Id     `json:"envId" gorm:"size:32;not null"`
	CreateAt       *Time  `json:"createAt" gorm:"type:datetime"`
	TaskId         Id     `json:"taskId" gorm:"size:32;not null"`
	Address        string `json:"address" gorm:"size:32;not null"`
	ResourceDetail []byte `json:"resourceDetail" gorm:"type:MEDIUMBLOB"`
}

func (ResourceDrift) TableName() string {
	return "iac_resource_drift"
}

// 偏移检测定时任务id和taskid 关系表
type EnvCronDrift struct {
	AutoUintIdModel
	EnvId       Id  `json:"envId" gorm:"size:32;not null"`
	CronEntryId int `json:"cronEntryId" gorm:"size:32;not null"`
}

func (EnvCronDrift) TableName() string {
	return "iac_env_cron_drift"
}

func (e *EnvCronDrift) Migrate(sess *db.Session) (err error) {
	if err = e.AddUniqueIndex(sess, "unique__task__cron__id",
		"env_id", "cron_entry_id"); err != nil {
		return err
	}
	return nil
}
