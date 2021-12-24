package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils/logs"
	"errors"
	"fmt"
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
	// 查询云模板对应的环境
	for _, tpl := range tplList {
		sysUserId := models.Id(consts.SysUserId)
		//todo 处理云模板的触发器
		envs, err := services.GetEnvByTplId(tx, tpl.Id)
		if err != nil {
			logs.Get().WithField("webhook", "searchEnv").
				Errorf("search env err: %v, tplId: %s", err, tpl.Id)
			// 记录个日志就行
			continue
		}

		for _, env := range envs {
			for _, v := range env.Triggers {
				var er error
				// 判断vcs类型，不同vcs, 入参不同
				switch vcs.VcsType {
				case consts.GitTypeGitLab:
					//er = gitlabActionPrOrPush(tx, v, sysUserId, &env, &tpl, form)
					er = actionPrOrPush(tx, v, sysUserId, &env, &tpl, form.Ref,
						form.ObjectAttributes.TargetBranch, form.ObjectAttributes.SourceBranch,
						form.ObjectAttributes.State, form.After, form.Before, form.ObjectAttributes.Iid)
				case consts.GitTypeGitEA:
					//er = giteaActionPrOrPush(tx, v, sysUserId, &env, &tpl, form)
					er = actionPrOrPush(tx, v, sysUserId, &env, &tpl, form.Ref,
						form.PullRequest.Base.Ref, form.PullRequest.Head.Ref,
						form.Action, form.After, form.Before, form.PullRequest.Number)
				case consts.GitTypeGithub:
					//er = githubActionPrOrPush(tx, v, sysUserId, &env, &tpl, form)
					er = actionPrOrPush(tx, v, sysUserId, &env, &tpl, form.Ref,
						form.PullRequest.Base.Ref, form.PullRequest.Head.Ref,
						form.Action, form.After, form.Before, form.PullRequest.Number)
				case consts.GitTypeGitee:
					//er = giteeActionPrOrPush(tx, v, sysUserId, &env, &tpl, form)
					er = actionPrOrPush(tx, v, sysUserId, &env, &tpl, form.Ref,
						form.PullRequest.Base.Ref, form.PullRequest.Head.Ref,
						form.Action, form.After, form.Before, form.PullRequest.Number)
				default:
					er = errors.New(fmt.Sprintf("vcs type error %s", vcs.VcsType))
				}

				if er != nil {
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
		Name: models.Task{}.GetTaskNameByType(taskType),

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

func actionPrOrPush(tx *db.Session, trigger string, userId models.Id,
	env *models.Env, tpl *models.Template, pushRef, baseRef, headRef, prStatus, afterCommit, beforeCommit string, prId int) error {

	// 比较分支
	// 如果同时不满足push分支和pr目标分支则不做动作
	if env.Revision != strings.Replace(pushRef, RefHeads, "", -1) &&
		env.Revision != baseRef {
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
