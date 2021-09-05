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
	"net/http"
	"strconv"
	"strings"
)

const (
	GitlabObjectKindPush = "push"
	GitlabPrOpened       = "opened"
	GitlabPrMerged       = "merged"
	RefHeads             = "refs/heads/"
)

func WebhooksApiHandler(c *ctx.ServiceContext, form forms.WebhooksApiHandler) (interface{}, e.Error) {

	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 根据VcsId & 仓库Id查询对应的云模板
	tplList, err := services.QueryTemplateByVcsIdAndRepoId(tx, form.VcsId, strconv.Itoa(int(form.Project.Id)))
	// 查询云模板对应的环境
	for _, tpl := range tplList {
		envs, err := services.GetEnvByTplId(tx, tpl.Id)
		if err != nil {
			logs.Get().WithField("webhook", "searchEnv").
				Errorf("search env err: %v, tplId: %s", err, tpl.Id)
			// 记录个日志就行
			continue
		}
		for _, env := range envs {
			for _, v := range env.Triggers {
				var err error

				// 比较分支
				if env.Revision != strings.Replace(form.Ref, RefHeads, "", -1) &&
					env.Revision != form.ObjectAttributes.TargetBranch {
					continue
				}
				// 判断pr类型并确认动作
				//open状态的mr进行plan计划
				if v == consts.EnvTriggerPRMR && form.ObjectAttributes.State == GitlabPrOpened {
					err = CreateWebhookTask(tx, models.TaskTypePlan, c.UserId, &env, &tpl)
				}

				if v == consts.EnvTriggerCommit && form.ObjectKind == GitlabObjectKindPush {
					err = CreateWebhookTask(tx, models.TaskTypeApply, c.UserId, &env, &tpl)
				}

				if err != nil {
					logs.Get().WithField("webhook", "createTask").
						Errorf("create task err: %v, envId: %s", err, env.Id)
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

func CreateWebhookTask(tx *db.Session, taskType string, userId models.Id, env *models.Env, tpl *models.Template) error {
	// 获取计算后的变量列表
	vars, err, _ := services.GetValidVariables(tx, consts.ScopeEnv, env.OrgId, env.ProjectId, env.TplId, env.Id, true)
	if err != nil {
		return e.New(err.Code(), err, http.StatusInternalServerError)
	}
	task := models.Task{
		Name:        models.Task{}.GetTaskNameByType(taskType),
		Type:        taskType,
		Flow:        models.TaskFlow{},
		Targets:     models.StrSlice{},
		CreatorId:   userId,
		KeyId:       env.KeyId,
		RunnerId:    env.RunnerId,
		Variables:   services.GetVariableBody(vars),
		StepTimeout: env.Timeout,
		AutoApprove: env.AutoApproval,
	}

	if _, err := services.CreateTask(tx, tpl, env, task); err != nil {
		_ = tx.Rollback()
		logs.Get().Errorf("error creating task, err %s", err)
		return e.New(err.Code(), err, http.StatusInternalServerError)
	}
	return nil
}
