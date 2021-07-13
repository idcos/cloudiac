package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"time"
)

func GetTaskSteps(sess *db.Session, taskId models.Id) ([]*models.TaskStep, error) {
	steps := make([]*models.TaskStep, 0)
	err := sess.Where(models.TaskStep{TaskId: taskId}).Order("index").Find(&steps)
	return steps, err
}

func GetTaskStep(sess *db.Session, taskId models.Id, step int) (*models.TaskStep, e.Error) {
	taskStep := models.TaskStep{}
	err := sess.Where(models.TaskStep{
		TaskId: taskId,
		Index:  step,
	}).First(&taskStep)

	if err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TaskStepNotExists)
		}
		return nil, e.New(e.DBError, err)
	}
	return &taskStep, nil
}

// ApproveTaskStep 标识步骤通过审批
func ApproveTaskStep(tx *db.Session, taskId models.Id, step int, userId models.Id) e.Error {
	if _, err := tx.Model(&models.TaskStep{}).
		Where("task_id = ? AND `index` = ?", taskId, step).
		Update(&models.TaskStep{ApproverId: userId}); err != nil {
		if e.IsRecordNotFound(err) {
			return e.New(e.TaskStepNotExists)
		}
		return e.New(e.DBError, err)
	}
	return nil
}

func UpdateTaskStep(sess *db.Session, taskStep *models.TaskStep) e.Error {
	if _, err := sess.Model(&models.TaskStep{}).Update(taskStep); err != nil {
		if e.IsRecordNotFound(err) {
			return e.New(e.TaskStepNotExists)
		}
		return e.New(e.DBError, err)
	}
	return nil
}

// RejectTaskStep 驳回步骤审批
func RejectTaskStep(dbSess *db.Session, taskId models.Id, step int, userId models.Id) e.Error {
	taskStep, er := GetTaskStep(dbSess, taskId, step)
	if er != nil {
		return e.AutoNew(er, e.DBError)
	}

	taskStep.ApproverId = userId
	return ChangeTaskStepStatus(dbSess, taskStep, models.TaskStepRejected, "")
}

func IsTerraformStep(typ string) bool {
	return utils.StrInArray(typ, models.TaskStepInit, models.TaskStepPlan,
		models.TaskStepApply, models.TaskStepDestroy)
}

// ChangeTaskStepStatus 修改步骤状态及 startAt、endAt，并同步修改任务状态
func ChangeTaskStepStatus(dbSess *db.Session, taskStep *models.TaskStep, status, message string) e.Error {
	taskStep.Status = status
	taskStep.Message = message

	now := time.Now()
	if taskStep.StartAt == nil && taskStep.IsStarted() {
		taskStep.StartAt = &now
	} else if taskStep.StartAt != nil && taskStep.EndAt == nil && taskStep.IsExited() {
		taskStep.EndAt = &now
	}

	logs.Get().WithField("taskId", taskStep.TaskId).
		WithField("step", taskStep.Index).
		Debugf("change step to '%s'", status)
	if _, err := dbSess.Model(&models.TaskStep{}).Update(taskStep); err != nil {
		return e.New(e.DBError, err)
	}

	if task, err := GetTask(dbSess, taskStep.TaskId); err != nil {
		return e.AutoNew(err, e.DBError)
	} else {
		return ChangeTaskStatusWithStep(dbSess, task, taskStep)
	}
}
