// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils/logs"
	"net/http"
	"strconv"
	"strings"
)

const (
	GitlabObjectKindPush = "push"
	GitlabPrOpened       = "opened"
	GitlabPrMerged       = "merged"
	RefHeads             = "refs/heads/"
	GiteePrOpen          = "open"
)

type webhookOptions struct {
	PushRef      string
	BaseRef      string
	HeadRef      string
	PrStatus     string
	AfterCommit  string
	BeforeCommit string
	PrId         int
}

func searchTplEnv(tx *db.Session, tplList []models.Template, options webhookOptions) {

	for tIndex, tpl := range tplList {
		sysUserId := models.Id(consts.SysUserId)

		if len(tpl.Triggers) > 0 {
			createTplScan(sysUserId, &tplList[tIndex], options)
		}

		envs, err := services.GetEnvByTplId(tx, tpl.Id)
		if err != nil {
			logs.Get().WithField("webhook", "searchEnv").
				Errorf("search env err: %v, tplId: %s", err, tpl.Id)
			// 记录个日志就行
			continue
		}

		for eIndex, env := range envs {
			// 跳过已归档环境
			if env.Archived {
				continue
			}
			for _, v := range env.Triggers {
				if er := actionPrOrPush(tx, v, sysUserId, &envs[eIndex], &tplList[tIndex], options); er != nil {
					logs.Get().WithField("webhook", "createTask").
						Errorf("create task er: %v, envId: %s", er, env.Id)
				}
			}
		}
	}
}

func WebhooksApiHandler(c *ctx.ServiceContext, form forms.WebhooksApiHandler) (interface{}, e.Error) {
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 查询vcs
	vcs, err := services.GetVcsById(tx, models.Id(form.VcsId))
	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("webhook get vcs err: %s", err)
		return nil, e.New(e.DBError, err)
	}

	// 根据VcsId & 仓库Id查询对应的云模板
	tplList, err := services.QueryTemplateByVcsIdAndRepoId(tx, form.VcsId, getVcsRepoId(vcs.VcsType, form))
	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("webhook get tpl err: %s", err)
		return nil, e.New(e.DBError, err)
	}
	options := webhookOptions{
		PushRef:      form.Ref,
		BaseRef:      form.PullRequest.Base.Ref,
		HeadRef:      form.PullRequest.Head.Ref,
		PrStatus:     form.Action,
		AfterCommit:  form.After,
		BeforeCommit: form.Before,
		PrId:         form.PullRequest.Number,
	}

	if vcs.VcsType == consts.GitTypeGitLab {
		options.BaseRef = form.ObjectAttributes.TargetBranch
		options.HeadRef = form.ObjectAttributes.SourceBranch
		options.PrStatus = form.ObjectAttributes.State
		options.PrId = form.ObjectAttributes.Iid
	}

	// 查询云模板对应的环境
	searchTplEnv(tx, tplList, options)

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error create task, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return nil, err
}

type CreateWebhookTaskParam struct {
	TaskType string
	Revision string
	CommitId string
	UserId   models.Id
	Env      *models.Env
	Tpl      *models.Template
	PrId     int
	Source   string
}

//nolint
func CreateWebhookTask(tx *db.Session, param CreateWebhookTaskParam) error {
	env := param.Env
	// 计算变量列表
	vars, er := services.GetValidVarsAndVgVars(tx, env.OrgId, env.ProjectId, env.TplId, env.Id)
	if er != nil {
		_ = tx.Rollback()
		return e.New(e.DBError, er, http.StatusInternalServerError)
	}
	task := &models.Task{
		Name:        models.Task{}.GetTaskNameByType(param.TaskType),
		Targets:     models.StrSlice{},
		CreatorId:   param.UserId,
		KeyId:       env.KeyId,
		Variables:   vars,
		AutoApprove: env.AutoApproval,
		Revision:    param.Revision,
		CommitId:    param.CommitId,
		BaseTask: models.BaseTask{
			Type:        param.TaskType,
			RunnerId:    env.RunnerId,
			StepTimeout: env.Timeout,
		},
		Source: param.Source,
	}

	task, err := services.CreateTask(tx, param.Tpl, env, *task)
	if err != nil {
		_ = tx.Rollback()
		logs.Get().Errorf("error creating task, err %s", err)
		return e.New(err.Code(), err, http.StatusInternalServerError)
	}

	if param.PrId != 0 && param.TaskType == models.TaskTypePlan {
		// 创建pr与作业的关系
		if err := services.CreateVcsPr(tx, models.VcsPr{
			PrId:   param.PrId,
			TaskId: task.Id,
			EnvId:  task.EnvId,
			VcsId:  param.Tpl.VcsId,
		}); err != nil {
			logs.Get().Errorf("error creating vcs pr, err %s", err)
			return e.New(err.Code(), err, http.StatusInternalServerError)
		}
	}
	logs.Get().Infof("create webhook task success. envId:%s, task type: %s", env.Id, param.TaskType)
	return nil
}

func checkVcsCallbackMessage(revision, pushRef, baseRef string) bool {
	// 比较分支
	// 如果同时不满足push分支和pr目标分支则不做动作
	if revision != strings.Replace(pushRef, RefHeads, "", -1) &&
		revision != baseRef {
		return false
	}
	return true
}

func actionPrOrPush(tx *db.Session, trigger string, userId models.Id,
	env *models.Env, tpl *models.Template, options webhookOptions) error {

	if !checkVcsCallbackMessage(env.Revision, options.PushRef, options.BaseRef) {
		logs.Get().WithField("webhook", "createTask").
			Infof("tplId: %s, envId: %s, revision don't match, env.revision: %s, %s or %s",
				env.TplId, env.Id, env.Revision, options.PushRef, options.BaseRef)
		return nil
	}

	// 判断pr类型并确认动作
	// open状态的mr进行plan计划
	if trigger == consts.EnvTriggerPRMR && options.PrStatus == GitlabPrOpened {
		// models.TaskTypePlan, options.HeadRef, "", userId, env, tpl, options.PrId, consts.TaskSourceWebhookPlan)
		param := CreateWebhookTaskParam{
			TaskType: models.TaskTypePlan,
			Revision: options.HeadRef,
			CommitId: "",
			UserId:   userId,
			Env:      env,
			Tpl:      tpl,
			PrId:     options.PrId,
			Source:   consts.TaskSourceWebhookPlan,
		}
		return CreateWebhookTask(tx, param)
	}
	// push操作，执行apply计划
	if trigger == consts.EnvTriggerCommit && options.BeforeCommit != "" {
		if env.LockedStatus {
			logs.Get().WithField("webhook", "createTask").Errorf("env %s is locked don't allow apply", env.Id)
			return nil
		}

		param := CreateWebhookTaskParam{
			TaskType: models.TaskTypeApply,
			Revision: env.Revision,
			CommitId: options.AfterCommit,
			UserId:   userId,
			Env:      env,
			Tpl:      tpl,
			PrId:     options.PrId,
			Source:   consts.TaskSourceWebhookApply,
		}
		return CreateWebhookTask(tx, param)
	}

	return nil
}

func getVcsRepoId(vcsType string, form forms.WebhooksApiHandler) string {
	switch vcsType {
	case consts.GitTypeGitLab:
		return strconv.Itoa(int(form.Project.Id))
	case consts.GitTypeGitEA:
		return strconv.Itoa(form.Repository.Id)
	case consts.GitTypeGithub:
		return form.Repository.FullName
	case consts.GitTypeGitee:
		return form.Repository.FullName
	default:
		return ""
	}
}

func createTplScan(userId models.Id, tpl *models.Template, options webhookOptions) {
	logger := logs.Get()
	// 云模板扫描未启用，不允许发起手动检测
	if enabled, err := services.IsTemplateEnabledScan(db.Get(), tpl.Id); err != nil {
		logger.Errorf("template enable err: %s", err)
		return
	} else if !enabled {
		logger.Infof("template %s not open scan", tpl.Id)
		return
	}

	if !checkVcsCallbackMessage(tpl.RepoRevision, options.PushRef, options.BaseRef) {
		return
	}

	// 目前云模板的webhook只有push一种
	if len(tpl.Triggers) > 0 && tpl.Triggers[0] != consts.EnvTriggerCommit {
		return
	}

	// 创建任务
	runnerId, err := services.GetDefaultRunnerId()
	if err != nil {
		logger.Errorf("webhook task scan get runner, err %s", err)
		return
	}

	tx := db.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	taskType := models.TaskTypeTplScan
	task, err := services.CreateScanTask(tx, tpl, nil, models.ScanTask{
		Name:      models.ScanTask{}.GetTaskNameByType(taskType),
		CreatorId: userId,
		TplId:     tpl.Id,
		BaseTask: models.BaseTask{
			Type:        taskType,
			StepTimeout: common.DefaultTaskStepTimeout,
			RunnerId:    runnerId,
		},
	})
	if err != nil {
		_ = tx.Rollback()
		logger.Errorf("error creating scan task, err %s", err)
		return
	}

	if err := services.InitScanResult(tx, task); err != nil {
		_ = tx.Rollback()
		logger.Errorf("task '%s' init scan result error: %v", task.Id, err)
		return
	}

	if task.Type == models.TaskTypeTplScan {
		tpl.LastScanTaskId = task.Id
		if _, err := db.Get().Save(tpl); err != nil {
			_ = tx.Rollback()
			logger.Errorf("save template, err %s", err)
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		logger.Errorf("commit env, err %s", err)
		return
	}
}
