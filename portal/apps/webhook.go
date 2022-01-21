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

	var (
		pushRef      string = form.Ref
		baseRef      string = form.PullRequest.Base.Ref
		headRef      string = form.PullRequest.Head.Ref
		prStatus     string = form.Action
		afterCommit  string = form.After
		beforeCommit string = form.Before
		prId         int    = form.PullRequest.Number
	)

	if vcs.VcsType == consts.GitTypeGitLab {
		baseRef = form.ObjectAttributes.TargetBranch
		headRef = form.ObjectAttributes.SourceBranch
		prStatus = form.ObjectAttributes.State
		prId = form.ObjectAttributes.Iid
	}

	// 查询云模板对应的环境
	for _, tpl := range tplList {
		sysUserId := models.Id(consts.SysUserId)

		if len(tpl.Triggers) > 0 {
			createTplScan(sysUserId, &tpl, pushRef, baseRef, headRef, prStatus, afterCommit, beforeCommit)
		}

		envs, err := services.GetEnvByTplId(tx, tpl.Id)
		if err != nil {
			logs.Get().WithField("webhook", "searchEnv").
				Errorf("search env err: %v, tplId: %s", err, tpl.Id)
			// 记录个日志就行
			continue
		}

		for _, env := range envs {
			// 跳过已归档环境
			if env.Archived {
				continue
			}
			for _, v := range env.Triggers {
				if er := actionPrOrPush(tx, v, sysUserId, &env, &tpl, pushRef, baseRef,
					headRef, prStatus, afterCommit, beforeCommit, prId); er != nil {
					logs.Get().WithField("webhook", "createTask").
						Errorf("create task er: %v, envId: %s", er, env.Id)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error create task, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return nil, err
}

func CreateWebhookTask(tx *db.Session, taskType, revision, commitId string, userId models.Id, env *models.Env, tpl *models.Template, prId int, source string) error {
	// 计算变量列表
	vars, er := services.GetValidVarsAndVgVars(tx, env.OrgId, env.ProjectId, env.TplId, env.Id)
	if er != nil {
		_ = tx.Rollback()
		return e.New(e.DBError, er, http.StatusInternalServerError)
	}

	task := &models.Task{
		Name:        models.Task{}.GetTaskNameByType(taskType),
		Targets:     models.StrSlice{},
		CreatorId:   userId,
		KeyId:       env.KeyId,
		Variables:   vars,
		AutoApprove: env.AutoApproval,
		Revision:    revision,
		CommitId:    commitId,
		BaseTask: models.BaseTask{
			Type:        taskType,
			RunnerId:    env.RunnerId,
			StepTimeout: env.Timeout,
		},
		Source: source,
	}

	task, err := services.CreateTask(tx, tpl, env, *task)
	if err != nil {
		_ = tx.Rollback()
		logs.Get().Errorf("error creating task, err %s", err)
		return e.New(err.Code(), err, http.StatusInternalServerError)
	}

	if prId != 0 && taskType == models.TaskTypePlan {
		// 创建pr与作业的关系
		if err := services.CreateVcsPr(tx, models.VcsPr{
			PrId:   prId,
			TaskId: task.Id,
			EnvId:  task.EnvId,
			VcsId:  tpl.VcsId,
		}); err != nil {
			logs.Get().Errorf("error creating vcs pr, err %s", err)
			return e.New(err.Code(), err, http.StatusInternalServerError)
		}
	}
	logs.Get().Infof("create webhook task success. envId:%s, task type: %s", env.Id, taskType)
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
	env *models.Env, tpl *models.Template, pushRef, baseRef, headRef, prStatus, afterCommit, beforeCommit string, prId int) error {

	if !checkVcsCallbackMessage(env.Revision, pushRef, baseRef) {
		logs.Get().WithField("webhook", "createTask").
			Infof("tplId: %s, envId: %s, revision don't match, env.revision: %s, %s or %s",
				env.TplId, env.Id, env.Revision, pushRef, baseRef)
		return nil
	}

	// 判断pr类型并确认动作
	// open状态的mr进行plan计划
	if trigger == consts.EnvTriggerPRMR && prStatus == GitlabPrOpened {
		return CreateWebhookTask(tx, models.TaskTypePlan, headRef, "", userId, env, tpl, prId, consts.TaskSourceWebhookPlan)
	}
	// push操作，执行apply计划
	if trigger == consts.EnvTriggerCommit && beforeCommit != "" {
		return CreateWebhookTask(tx, models.TaskTypeApply, env.Revision, afterCommit, userId, env, tpl, prId, consts.TaskSourceWebhookApply)
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

func createTplScan(userId models.Id, tpl *models.Template, pushRef, baseRef, headRef, prStatus, afterCommit, beforeCommit string) {
	logger := logs.Get()
	// 云模板扫描未启用，不允许发起手动检测
	if enabled, err := services.IsTemplateEnabledScan(db.Get(), tpl.Id); err != nil {
		logger.Errorf("template enable err: %s", err)
		return
	} else if !enabled {
		logger.Infof("template %s not open scan", tpl.Id)
		return
	}

	if !checkVcsCallbackMessage(tpl.RepoRevision, pushRef, baseRef) {
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

	taskType := models.TaskTypeTplScan
	task, err := services.CreateScanTask(db.Get(), tpl, nil, models.ScanTask{
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
		logger.Errorf("error creating scan task, err %s", err)
		return
	}

	if task.Type == models.TaskTypeTplScan {
		tpl.LastScanTaskId = task.Id
		if _, err := db.Get().Save(tpl); err != nil {
			logger.Errorf("save template, err %s", err)
		}
	}
}
