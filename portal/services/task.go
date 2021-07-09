package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
	"github.com/pkg/errors"
)

func GetTask(dbSess *db.Session, id models.Id) (*models.Task, e.Error) {
	task := models.Task{}
	err := dbSess.Where("id = ?", id).First(&task)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TaskNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &task, nil
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

	for i, step := range task.Flow.Steps {
		if len(task.Targets) != 0 && IsTerraformStep(step.Type) {
			if step.Type != models.TaskStepInit {
				for _, t := range task.Targets {
					step.Args = append(step.Args, fmt.Sprintf("-target=%s", t))
				}
			}
			// TODO: tfVars, playVars 也以这种方式传入？
		}

		if _, er := createTaskStep(tx, task, step, i); er != nil {
			return nil, e.New(er.Code(), errors.Wrapf(er, "save task step"))
		}
	}
	return &task, nil
}

func createTaskStep(tx *db.Session, task models.Task, stepBody models.TaskStepBody, index int) (*models.TaskStep, e.Error) {
	s := models.TaskStep{
		TaskStepBody: stepBody,
		OrgId:        task.OrgId,
		ProjectId:    task.ProjectId,
		TaskId:       task.Id,
		Index:        index,
		Status:       models.TaskStepPending,
		Message:      "",
	}
	s.Id = models.NewId("step")
	s.LogPath = s.GenLogPath()

	if _, err := tx.Save(&s); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return &s, nil
}

func GetTaskById(tx *db.Session, id models.Id) (*models.Task, e.Error) {
	o := models.Task{}
	if err := tx.Where("id = ?", id).First(&o); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TaskNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &o, nil
}

func QueryTask(query *db.Session) *db.Session {
	return query.Model(&models.Task{})
}
