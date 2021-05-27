package utils

import (
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/utils/logs"
	"github.com/jinzhu/gorm"
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
		Where("name = 'MAX_JOBS_PER_RUNNER'").First(&systemCfg); err != nil && err != gorm.ErrRecordNotFound {
		logger.Debugf("db err: %v", err)
	}
	if Str2int(systemCfg.Value) > 0 {
		UpdateRunnerMax(Str2int(systemCfg.Value))
	}

}
