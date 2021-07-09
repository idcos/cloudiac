package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
)

func GetTaskSteps(sess *db.Session, taskId models.Id) ([]*models.TaskStep, error) {
	steps := make([]*models.TaskStep, 0)
	err := sess.Where(models.TaskStep{TaskId: taskId}).Order("index").Find(&steps)
	return steps, err
}

func GetTaskStep(sess *db.Session, taskId models.Id, step int) (*models.TaskStep, error) {
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
		Where("task_id = ? AND index = ?", taskId, step).
		Update(&models.TaskStep{ApproverId: userId}); err != nil {
		if e.IsRecordNotFound(err) {
			return e.New(e.TaskStepNotExists)
		}
		return e.New(e.DBError, err)
	}
	return nil
}

func IsTerraformStep(typ string) bool {
	return utils.StrInArray(typ, models.TaskStepInit, models.TaskStepPlan,
		models.TaskStepApply, models.TaskStepDestroy)
}
