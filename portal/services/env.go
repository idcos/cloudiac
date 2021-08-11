// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils/logs"
	"fmt"
	"time"
)

func GetEnv(sess *db.Session, id models.Id) (*models.Env, error) {
	env := models.Env{}
	err := sess.Where("id = ?", id).First(&env)
	return &env, err
}

func CreateEnv(tx *db.Session, env models.Env) (*models.Env, e.Error) {
	if env.Id == "" {
		env.Id = models.NewId("env")
	}
	if env.StatePath == "" {
		env.StatePath = env.DefaultStatPath()
	}
	if err := models.Create(tx, &env); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.EnvAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &env, nil
}

func UpdateEnv(tx *db.Session, id models.Id, attrs models.Attrs) (env *models.Env, re e.Error) {
	env = &models.Env{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Env{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.EnvAliasDuplicate)
		}
		return nil, e.New(e.DBError, fmt.Errorf("update env error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(env); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query env error: %v", err))
	}
	return
}

func UpdateEnvModel(tx *db.Session, id models.Id, env models.Env) e.Error {
	_, err := models.UpdateModel(tx.Where("id = ?", id), &env)
	if err != nil {
		return e.AutoNew(err, e.DBError)
	}
	return nil
}

func DeleteEnv(tx *db.Session, id models.Id) e.Error {
	if _, err := tx.Where("id = ?", id).Delete(&models.Env{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete env error: %v", err))
	}
	return nil
}

func GetEnvById(tx *db.Session, id models.Id) (*models.Env, e.Error) {
	o := models.Env{}
	if err := tx.Model(models.Env{}).Where("id = ?", id).First(&o); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.EnvNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &o, nil
}

func QueryEnvDetail(query *db.Session) *db.Session {
	query = query.Model(&models.Env{})

	// 模板名称
	query = query.Joins("left join iac_template as t on t.id = iac_env.tpl_id").
		LazySelectAppend("t.name as template_name,iac_env.*")
	// 创建人姓名
	query = query.Joins("left join iac_user as u on u.id = iac_env.creator_id").
		LazySelectAppend("u.name as creator,iac_env.*")
	// 资源数量统计
	query = query.Joins("left join (select count(*) as resource_count, task_id from iac_resource group by task_id) as r on r.task_id = iac_env.last_task_id").
		LazySelectAppend("r.resource_count, iac_env.*")
	// 密钥名称
	query = query.Joins("left join iac_key as k on k.id = iac_env.key_id").
		LazySelectAppend("k.name as key_name,iac_env.*")

	return query
}

func GetEnvDetailById(query *db.Session, id models.Id) (*models.EnvDetail, e.Error) {
	d := models.EnvDetail{}
	if err := query.Where("iac_env.id = ?", id).First(&d); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.EnvNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &d, nil
}

func GetEnvByTplId(tx *db.Session, id models.Id) ([]models.Env, error) {
	env := make([]models.Env, 0)
	if err := tx.Where("tpl_id = ?", id).Find(&env); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return env, nil
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
		case models.TaskRejected:
			// 任务驳回，环境状态不变
			break
		case models.TaskFailed:
			envStatus = models.EnvStatusFailed
		case models.TaskComplete:
			if task.Type == models.TaskTypeApply {
				envStatus = models.EnvStatusActive
			} else if task.Type == models.TaskTypeDestroy {
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
		logger.Infof("change env to '%v'", envStatus)
		attrs["status"] = envStatus
	}
	_, err := tx.Model(&models.Env{}).Where("id = ?", id).UpdateAttrs(attrs)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return e.New(e.EnvNotExists)
		}
		return e.New(e.DBError, err)
	}
	return nil
}

func GetVariableBody(vars models.EnvVariables) []models.VariableBody {
	var vb []models.VariableBody
	for _, v := range vars {
		vb = append(vb, models.VariableBody{
			Scope:       v.Scope,
			Type:        v.Type,
			Name:        v.Name,
			Value:       v.Value,
			Sensitive:   v.Sensitive,
			Description: v.Description,
		})
	}
	return vb
}

var (
	ttlMap = map[string]string{
		"1d":  "24h",
		"3d":  "72h",
		"1w":  "168h",
		"15d": "360h",
		"30d": "720h",
	}
)

func ParseTTL(ttl string) (time.Duration, error) {
	ds, ok := ttlMap[ttl]
	if ok {
		return time.ParseDuration(ds)
	}
	// map 中不存在则尝试直接解析
	t, err := time.ParseDuration(ttl)
	if err != nil {
		return t, fmt.Errorf("invalid duration: %v", ttl)
	}
	return t, nil
}
