// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils/logs"
)

func CreateAutoDeployTask(tx *db.Session, env *models.Env) (*models.Task, e.Error) {
	logger := logs.Get()

	tpl, err := GetTemplateById(tx, env.TplId)
	if err != nil {
		logger.Errorf("get template %s error: %v", env.TplId, err)
		return nil, err
	}
	// 计算变量列表
	vars, er := GetValidVarsAndVgVars(tx, env.OrgId, env.ProjectId, env.TplId, env.Id)
	if er != nil {
		return nil, err
	}

	paramTask := models.Task{
		Name:            consts.TaskAutoDeployName,
		Targets:         env.Targets,
		CreatorId:       consts.SysUserId,
		KeyId:           env.KeyId,
		Variables:       vars,
		AutoApprove:     true,
		Revision:        env.Revision,
		StopOnViolation: env.StopOnViolation,
		ExtraData:       env.ExtraData,
		BaseTask: models.BaseTask{
			Type:        models.TaskTypeApply,
			StepTimeout: env.StepTimeout,
			RunnerId:    env.RunnerId,
		},
		Source: consts.TaskSourceAutoDeploy,
	}

	if env.LastResTaskId != "" {
		// 自动部署任务使用环境最后一次部署时的 pipeline
		lastResTask, err := GetTaskById(tx, env.LastResTaskId)
		if err != nil {
			logger.Errorf("get env lastResTask error: %v", err)
			return nil, err
		}
		paramTask.Pipeline = lastResTask.Pipeline
		paramTask.CommitId = lastResTask.CommitId
	}

	return CreateTask(tx, tpl, env, paramTask)
}
