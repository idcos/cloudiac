// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"cloudiac/utils/logs"
)

var runnerMax int

func GetRunnerMax() int {
	return runnerMax
}

func UpdateRunnerMax(max int) {
	runnerMax = max
}

func MaintenanceRunnerPerMax() {
	logger := logs.Get().WithField("action", "MaintenanceRunnerPerMax")
	systemCfg := models.SystemCfg{}
	if err := db.Get().Table(models.SystemCfg{}.TableName()).
		Where("name = 'MAX_JOBS_PER_RUNNER'").First(&systemCfg); err != nil && e.IsRecordNotFound(err) {
		logger.Debugf("db err: %v", err)
	}

	if utils.Str2int(systemCfg.Value) > 0 {
		UpdateRunnerMax(utils.Str2int(systemCfg.Value))
	}
}
