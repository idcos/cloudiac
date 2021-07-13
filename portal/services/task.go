package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils/logs"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

func GetTask(dbSess *db.Session, id models.Id) (*models.Task, e.Error) {
	task := models.Task{}
	err := dbSess.Where("id = ?", id).First(&task)
	if e.IsRecordNotFound(err) {
		return nil, e.New(e.TaskNotExists)
	}
	return &task, e.New(e.DBError, err)
}

func CreateTask(tx *db.Session, env *models.Env, p models.Task) (*models.Task, e.Error) {
	task := models.Task{
		// 以下为需要外部传入的属性
		Name:        p.Name,
		Type:        p.Type,
		Flow:        p.Flow,
		Targets:     p.Targets,
		CommitId:    p.CommitId,
		CreatorId:   p.CreatorId,
		RunnerId:    p.RunnerId,
		Variables:   p.Variables,
		StepTimeout: p.StepTimeout,
		AutoApprove: p.AutoApprove,

		OrgId:     env.OrgId,
		ProjectId: env.ProjectId,
		TplId:     env.TplId,
		EnvId:     env.Id,
		Status:    models.TaskPending,
		Message:   "",
		CurrStep:  0,
	}

	task.Id = models.NewId("run")
	if task.RunnerId == "" {
		task.RunnerId = env.RunnerId
	}

	if len(task.Flow.Steps) == 0 {
		var err error
		task.Flow, err = models.DefaultTaskFlow(task.Type)
		if err != nil {
			return nil, e.New(e.InternalError, err)
		}
	}

	if _, err := tx.Save(&task); err != nil {
		return nil, e.New(e.DBError, errors.Wrapf(err, "save task"))
	}

	taskSteps := make([]*models.TaskStep, 0)
	for i, step := range task.Flow.Steps {
		if len(task.Targets) != 0 && IsTerraformStep(step.Type) {
			if step.Type != models.TaskStepInit {
				for _, t := range task.Targets {
					step.Args = append(step.Args, fmt.Sprintf("-target=%s", t))
				}
			}
			// TODO: tfVars, playVars 也以这种方式传入？
		}

		taskStep, er := createTaskStep(tx, task, step, i)
		if er != nil {
			return nil, e.New(er.Code(), errors.Wrapf(er, "save task step"))
		}
		taskSteps = append(taskSteps, taskStep)
	}

	for i := range taskSteps {
		step := taskSteps[i]
		if i < len(taskSteps)-1 {
			step.NextStep = taskSteps[i+1].Id
		}
		if _, err := tx.Model(&models.TaskStep{}).Update(step); err != nil {
			return nil, e.New(e.DBError, err)
		}
	}

	return &task, nil
}

var stepStatus2TaskStatus = map[string]string{
	models.TaskStepPending:   models.TaskPending,
	models.TaskStepApproving: models.TaskApproving,
	models.TaskStepRejected:  models.TaskFailed,
	models.TaskStepRunning:   models.TaskRunning,
	models.TaskStepFailed:    models.TaskFailed,
	models.TaskStepTimeout:   models.TaskFailed,
	// complete 状态需要特殊处理
}

func ChangeTaskStatusWithStep(dbSess *db.Session, task *models.Task, step *models.TaskStep) e.Error {
	var (
		taskStatus string
	)

	if step.Status == models.TaskStepComplete {
		if step.NextStep == "" {
			taskStatus = models.TaskComplete
		} else {
			taskStatus = models.TaskRunning
		}
	} else {
		var ok bool
		taskStatus, ok = stepStatus2TaskStatus[step.Status]
		if !ok {
			panic(fmt.Errorf("unknown task step status %v", step.Status))
		}
	}

	return ChangeTaskStatus(dbSess, task, taskStatus, step.Message)
}

// ChangeTaskStatus 修改任务状态(同步修改 StartAt、EndAt 等)，并同步修改 env 状态
func ChangeTaskStatus(dbSess *db.Session, task *models.Task, status, message string) e.Error {
	task.Status = status
	task.Message = message

	now := time.Now()
	if task.StartAt == nil && task.Started() {
		task.StartAt = &now
	}
	if task.EndAt == nil && task.Exited() {
		task.EndAt = &now
	}

	logs.Get().WithField("taskId", task.Id).Debugf("change task to '%s'", status)
	if _, err := dbSess.Model(&models.Task{}).Update(task); err != nil {
		return e.AutoNew(err, e.DBError)
	}

	step, er := GetTaskStep(dbSess, task.Id, task.CurrStep)
	if er != nil {
		return e.AutoNew(er, e.DBError)
	}
	return ChangeEnvStatusWithTaskAndStep(dbSess, task.EnvId, task, step)
}
