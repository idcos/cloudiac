package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"github.com/pkg/errors"
)

func GetTask(dbSess *db.Session, id models.Id) (*models.Task, error) {
	task := models.Task{}
	err := dbSess.Where("id = ?", id).First(&task)
	return &task, err
}

func CreateTask(tx *db.Session, env *models.Env, p models.Task) (*models.Task, e.Error) {
	task := models.Task{
		CreatorId: p.CreatorId,
		RunnerId:  p.RunnerId,
		CommitId:  p.CommitId,
		Timeout:   p.Timeout,
		Type:      p.Type,
		Name:      p.Name,
		Flow:      p.Flow,
		Variables: p.Variables,
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
