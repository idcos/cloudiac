package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils/logs"
	"fmt"
)

func GetEnv(sess *db.Session, id models.Id) (*models.Env, error) {
	env := models.Env{}
	err := sess.Where("id = ?", id).First(&env)
	return &env, err
}

// ChangeEnvStatusWithTaskAndStep 基于任务和步骤的状态更新环境状态
func ChangeEnvStatusWithTaskAndStep(tx *db.Session, id models.Id, task *models.Task, step *models.TaskStep) e.Error {
	var (
		envStatus     = ""
		envTaskStatus = ""
		isDeploying   = false
	)

	// 不修改环境数据的任务也不会影响环境状态
	if !task.IsEffectTask() {
		return nil
	}

	if task.Exited() {
		switch task.Status {
		case models.TaskFailed:
			if step.Status != models.TaskStepRejected {
				envStatus = models.EnvStatusFailed
			}
		case models.TaskComplete:
			if task.Type == models.TaskTypeApply {
				envStatus = models.EnvStatusActive
			} else { // destroy 任务
				envStatus = models.EnvStatusInactive
			}
		default:
			return e.New(e.InternalError, fmt.Errorf("unknown exited task status: %v", task.Status))
		}
	} else if task.Started() {
		envTaskStatus = task.Status
		isDeploying = true
	} else { // pending
		// 任务进入 pending 状态不修改环境状态， 因为任务 pending 时可能同一个环境的其他任务正在执行
		// (实际目前任务创建后即进入 pending 状态，并不触发 change status 调用链)
		return nil
	}

	logger := logs.Get().WithField("envId", id)
	attrs := models.Attrs{
		"task_status": envTaskStatus,
		"deploying":   isDeploying,
	}
	if envStatus != "" {
		logger.Debugf("change env to '%v'", envStatus)
		attrs["status"] = envStatus
	}
	_, err := tx.Model(&models.Env{}).Where("id = ?", id).Update(attrs)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return e.New(e.EnvNotExists)
		}
		return e.New(e.DBError, err)
	}
	return nil
}
