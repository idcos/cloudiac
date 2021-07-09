package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
	"github.com/pkg/errors"
)

func GetTask(dbSess *db.Session, id models.Id) (*models.Task, error) {
	task := models.Task{}
	err := dbSess.Where("id = ?", id).First(&task)
	return &task, err
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
