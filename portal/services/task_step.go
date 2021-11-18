// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/common"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"time"
)

func GetTaskSteps(sess *db.Session, taskId models.Id) ([]*models.TaskStep, error) {
	steps := make([]*models.TaskStep, 0)
	err := sess.Where(models.TaskStep{TaskId: taskId}).Order("`index`").Find(&steps)
	return steps, err
}

func GetTaskStep(sess *db.Session, taskId models.Id, step int) (*models.TaskStep, e.Error) {
	taskStep := models.TaskStep{}
	err := sess.Where("task_id = ? AND `index` = ?", taskId, step).First(&taskStep)

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

	taskStep, er := GetTaskStep(tx, taskId, step)
	if er != nil {
		return e.AutoNew(er, e.DBError)
	}

	task, err := GetTask(tx, taskId)
	if err != nil {
		return err
	}

	// 审批通过将步骤标识为 pending 状态，任务被同步修改为 running 状态，
	// task manager 会在检测到步骤通过审批后开始执行步骤, 并标识为 running 状态
	return ChangeTaskStepStatusAndUpdate(tx, task, taskStep, models.TaskStepPending, "")
}

// RejectTaskStep 驳回步骤审批
func RejectTaskStep(dbSess *db.Session, taskId models.Id, step int, userId models.Id) e.Error {
	taskStep, er := GetTaskStep(dbSess, taskId, step)
	if er != nil {
		return e.AutoNew(er, e.DBError)
	}

	taskStep.ApproverId = userId

	if task, err := GetTask(dbSess, taskStep.TaskId); err != nil {
		return e.AutoNew(err, e.DBError)
	} else {
		return ChangeTaskStepStatusAndUpdate(dbSess, task, taskStep, models.TaskStepRejected, "rejected")
	}
}

func IsTerraformStep(typ string) bool {
	return utils.StrInArray(typ, models.TaskStepInit, models.TaskStepPlan,
		models.TaskStepApply, models.TaskStepDestroy)
}

func ChangeTaskStepStatusAndExitCode(dbSess *db.Session, task models.Tasker, taskStep *models.TaskStep,
	status, message string, exitCode int) e.Error {
	taskStep.ExitCode = exitCode
	return ChangeTaskStepStatusAndUpdate(dbSess, task, taskStep, status, message)
}

// ChangeTaskStepStatusAndUpdate 修改步骤状态并更新 taskStep
// 该函数会同步修改任务状态
func ChangeTaskStepStatusAndUpdate(dbSess *db.Session, task models.Tasker, taskStep *models.TaskStep, status, message string) e.Error {
	if taskStep.Status == status && message == "" {
		return nil
	}

	taskStep.Status = status
	taskStep.Message = message

	now := models.Time(time.Now())
	if taskStep.StartAt == nil && taskStep.IsStarted() {
		taskStep.StartAt = &now
	} else if taskStep.StartAt != nil && taskStep.EndAt == nil && taskStep.IsExited() {
		taskStep.EndAt = &now
	}

	if taskStep.Id == "" {
		// id 为空表示是生成的功能性步骤，非任务的流程步骤，
		// 不需要保存到 db，且步骤的状态不影响任务和环境的状态
		return nil
	}

	logger := logs.Get().WithField("taskId", taskStep.TaskId).WithField("step", taskStep.Index)
	if message != "" {
		logger.Infof("change step to '%s', message: %s", status, message)
	} else {
		logger.Debugf("change step to '%s'", status)
	}

	if _, err := dbSess.Model(taskStep).Update(taskStep); err != nil {
		return e.New(e.DBError, err)
	}

	if taskStep.IsExited() && !taskStep.IsRejected() {
		// 步骤结束时任务不能同步修改状态，需要等资源采集步骤执行结束并生成统计数据后才能更新任务状态。
		// 特殊的: 审批驳回的任务执行结束后不需要进行资源统计，应该立即修改状态
		return nil
	}

	// callback 步骤不影响任务状态
	if taskStep.IsCallback {
		return nil
	}

	return ChangeTaskStatusWithStep(dbSess, task, taskStep)
}

func newTaskStep(tx *db.Session, task models.Task, stepBody models.PipelineStep, index int) *models.TaskStep {
	s := models.TaskStep{
		PipelineStep: stepBody,
		OrgId:        task.OrgId,
		ProjectId:    task.ProjectId,
		EnvId:        task.EnvId,
		TaskId:       task.Id,
		Index:        index,
		Status:       models.TaskStepPending,
		Message:      "",
		NextStep:     "",
		RetryNumber:  task.RetryNumber,
	}

	// apply 和 destroy 步骤需要审批
	if !task.AutoApprove && (s.Type == common.TaskStepTfApply || s.Type == common.TaskStepTfDestroy) {
		s.MustApproval = true
	}

	s.Id = models.NewId("step")
	s.LogPath = s.GenLogPath()
	return &s
}

func newScanTaskStep(tx *db.Session, task models.ScanTask, stepBody models.PipelineStep, index int) *models.TaskStep {
	s := models.TaskStep{
		PipelineStep: stepBody,
		OrgId:        task.OrgId,
		TaskId:       task.Id,
		Index:        index,
		Status:       models.TaskStepPending,
		Message:      "",
	}
	s.Id = models.NewId("step")
	s.LogPath = s.GenLogPath()

	return &s
}

func GetTaskScanStep(query *db.Session, taskId models.Id) (*models.TaskStep, e.Error) {
	taskStep := models.TaskStep{}
	err := query.Where("task_id = ? AND `type` = ?", taskId, common.TaskStepOpaScan).First(&taskStep)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TaskStepNotExists)
		}
		return nil, e.New(e.DBError, err)
	}
	return &taskStep, nil
}

func GetTaskLastStep(sess *db.Session, taskId models.Id) (*models.TaskStep, e.Error) {
	taskStep := models.TaskStep{}
	err := sess.Where("task_id = ?", taskId).Order("`index` DESC").Limit(1).Find(&taskStep)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TaskStepNotExists)
		}
		return nil, e.New(e.DBError, err)
	}
	return &taskStep, nil
}

func CreateTaskStep(tx *db.Session, task models.Task, stepBody models.PipelineStep, index int) (*models.TaskStep, e.Error) {
	step := newTaskStep(tx, task, stepBody, index)
	if err := tx.Insert(step); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return step, nil
}

func CreateTaskCallbackStep(sess *db.Session, task models.Task, stepBody models.PipelineStep, index int) (*models.TaskStep, e.Error) {
	step := newTaskStep(sess, task, stepBody, index)
	step.IsCallback = true
	if err := sess.Insert(step); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return step, nil
}

func GetTaskPlanStep(sess *db.Session, taskId models.Id) (*models.TaskStep, e.Error) {
	taskStep := models.TaskStep{}
	err := sess.Where("task_id = ?", taskId).
		Where("type = ?", common.TaskStepTfPlan).First(&taskStep)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TaskStepNotExists)
		}
		return nil, e.New(e.DBError, err)
	}
	return &taskStep, nil
}
