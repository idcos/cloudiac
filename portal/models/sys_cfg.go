// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/portal/libs/db"
)

const (
	SysCfgNameMaxJobsPerRunner = "MAX_JOBS_PER_RUNNER"
	SysCfgNamePeriodOfLogSave  = "PERIOD_OF_LOG_SAVE"
	SysCfgNamRegistryAddr      = "REGISTRY_ADDR"
	SysCfgNameTaskStepTimeout  = "TASK_STEP_TIMEOUT"
)

type SystemCfg struct {
	BaseModel

	Name        string `json:"name" gorm:"not null;comment:设定名"`
	Value       string `json:"value" gorm:"not null;comment:设定值"`
	Description string `json:"description" gorm:"comment:描述"`
}

func (SystemCfg) TableName() string {
	return "iac_system_cfg"
}

func (o SystemCfg) Migrate(sess *db.Session) (err error) {
	if err := o.AddUniqueIndex(sess,
		"unique__system_cfg__name", "name"); err != nil {
		return err
	}

	if err := sess.ModifyModelColumn(&SystemCfg{}, "value"); err != nil {
		return err
	}

	if err := sess.ModifyModelColumn(&SystemCfg{}, "description"); err != nil {
		return err
	}

	return nil
}
