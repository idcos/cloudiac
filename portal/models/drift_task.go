package models

import "cloudiac/portal/libs/db"

type CronDriftTask struct {
	AutoUintIdModel
	EnvId          Id     `json:"envId" gorm:"size:32;not null"`
	CreateTime     *Time  `json:"createTime" gorm:"type:datetime"`
	TaskId         Id     `json:"taskId" gorm:"size:32;not null"`
	Address        string `json:"address" gorm:"size:32;not null"`
	ResourceDetail []byte `json:"resourceDetail" gorm:"type:MEDIUMBLOB"`
}

func (CronDriftTask) TableName() string {
	return "iac_cron_drift_task"
}

func (c *CronDriftTask) Migrate(sess *db.Session) (err error) {
	return TaskModelMigrate(sess, c)
}
