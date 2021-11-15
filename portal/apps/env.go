// Copyright 2021 CloudJ Company Limited. All rights reserved.

package apps

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/portal/services/vcsrv"
	"cloudiac/utils"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
)

// CreateEnv 创建环境
func CreateEnv(c *ctx.ServiceContext, form *forms.CreateEnvForm) (*models.EnvDetail, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create env %s", form.Name))

	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	// 检查模板
	query := c.DB().Where("status = ?", models.Enable)
	tpl, err := services.GetTemplateById(query, form.TplId)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get template, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 以下值只在未传入时使用模板定义的值，如果入参有该字段即使值为空也不会使用模板中的值
	if !form.HasKey("tfVarsFile") {
		form.TfVarsFile = tpl.TfVarsFile
	}
	if !form.HasKey("playVarsFile") {
		form.PlayVarsFile = tpl.PlayVarsFile
	}
	if !form.HasKey("playbook") {
		form.Playbook = tpl.Playbook
	}

	if form.Timeout == 0 {
		form.Timeout = common.DefaultTaskStepTimeout
	}

	var (
		destroyAt models.Time
	)

	if form.DestroyAt != "" {
		var err error
		destroyAt, err = models.Time{}.Parse(form.DestroyAt)
		if err != nil {
			return nil, e.New(e.BadParam, http.StatusBadRequest, err)
		}
	} else if form.TTL != "" {
		_, err := services.ParseTTL(form.TTL) // 检查 ttl 格式
		if err != nil {
			return nil, e.New(e.BadParam, http.StatusBadRequest, err)
		}
	}

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	var runnerId string = form.RunnerId
	if runnerId == "" {
		rId, err := services.GetDefaultRunner()
		if err != nil {
			return nil, err
		}
		runnerId = rId
	}

	env, err := services.CreateEnv(tx, models.Env{
		OrgId:     c.OrgId,
		ProjectId: c.ProjectId,
		CreatorId: c.UserId,
		TplId:     form.TplId,

		Name:     form.Name,
		RunnerId: runnerId,
		Status:   models.EnvStatusInactive,
		OneTime:  form.OneTime,
		Timeout:  form.Timeout,

		// 模板参数
		TfVarsFile:   form.TfVarsFile,
		PlayVarsFile: form.PlayVarsFile,
		Playbook:     form.Playbook,
		Revision:     form.Revision,
		KeyId:        form.KeyId,

		TTL:             form.TTL,
		AutoDestroyAt:   &destroyAt,
		AutoApproval:    form.AutoApproval,
		StopOnViolation: form.StopOnViolation,

		Triggers:    form.Triggers,
		RetryAble:   form.RetryAble,
		RetryDelay:  form.RetryDelay,
		RetryNumber: form.RetryNumber,

		ExtraData: models.JSON(form.ExtraData),
		Callback:  form.Callback,
	})
	if err != nil && err.Code() == e.EnvAlreadyExists {
		_ = tx.Rollback()
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating env, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	// sampleVariables是外部下发的环境变量，会和计算出来的变量列表冲突，这里需要做一下处理
	// FIXME 未对变量组的变量进行处理
	sampleVars, err := services.GetSampleValidVariables(tx, c.OrgId, c.ProjectId, env.TplId, env.Id, form.SampleVariables)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	if len(sampleVars) > 0 {
		form.Variables = append(form.Variables, sampleVars...)
	}

	// 创建新导入的变量
	if err = services.OperationVariables(tx, c.OrgId, c.ProjectId, env.TplId, env.Id, form.Variables, nil); err != nil {
		_ = tx.Rollback()
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	// 创建变量组与实例的关系
	if err := services.BatchUpdateRelationship(tx, form.VarGroupIds, form.DelVarGroupIds, consts.ScopeEnv, env.Id.String()); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	targets := make([]string, 0)
	if len(strings.TrimSpace(form.Targets)) > 0 {
		targets = strings.Split(strings.TrimSpace(form.Targets), ",")
	}

	// 计算变量列表
	vars, er := services.GetValidVarsAndVgVars(tx, env.OrgId, env.ProjectId, env.TplId, env.Id)
	if er != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// 创建任务
	task, err := services.CreateTask(tx, tpl, env, models.Task{
		Name:            models.Task{}.GetTaskNameByType(form.TaskType),
		Targets:         targets,
		CreatorId:       c.UserId,
		KeyId:           env.KeyId,
		Variables:       vars,
		AutoApprove:     env.AutoApproval,
		Revision:        env.Revision,
		StopOnViolation: env.StopOnViolation,
		BaseTask: models.BaseTask{
			Type:        form.TaskType,
			StepTimeout: form.Timeout,
			RunnerId:    runnerId,
		},
		ExtraData: models.JSON(form.ExtraData),
		Callback:  form.Callback,
	})
	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating task, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	// 首次部署，直接更新 last_task_id
	env.LastTaskId = task.Id
	if _, err := tx.Save(env); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error save env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 创建完成
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit env, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	envDetail := models.EnvDetail{
		Env:        *env,
		TaskId:     task.Id,
		Operator:   c.Username,
		OperatorId: c.UserId,
	}
	vcs, _ := services.QueryVcsByVcsId(tpl.VcsId, c.DB())
	if err := vcsrv.SetWebhook(vcs, tpl.RepoId, form.Triggers); err != nil {
		c.Logger().Errorf("set webhook err :%v", err)
	}
	return &envDetail, nil
}

// SearchEnv 环境查询
func SearchEnv(c *ctx.ServiceContext, form *forms.SearchEnvForm) (interface{}, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	query := c.DB().Where("iac_env.org_id = ? AND iac_env.project_id = ?", c.OrgId, c.ProjectId)
	query = services.QueryEnvDetail(query)

	if form.Status != "" {
		if utils.InArrayStr(models.EnvStatus, form.Status) {
			query = query.Where("iac_env.status = ? and iac_env.deploying = 0", form.Status)
		} else if utils.InArrayStr(models.EnvTaskStatus, form.Status) {
			query = query.Where("iac_env.task_status = ? and iac_env.deploying = 1", form.Status)
		} else {
			return nil, e.New(e.BadParam, http.StatusBadRequest)
		}
	}

	// 环境归档状态
	switch form.Archived {
	case "":
		// 默认返回未归档环境
		query = query.Where("iac_env.archived = ?", 0)
	case "all":
	// do nothing
	case "true":
		// 已归档
		query = query.Where("iac_env.archived = 1")
	case "false":
		// 未归档
		query = query.Where("iac_env.archived = 0")
	default:
		return nil, e.New(e.BadParam, http.StatusBadRequest)
	}

	if form.Q != "" {
		query = query.WhereLike("iac_env.name", form.Q)
	}

	// 默认按创建时间逆序排序
	if form.SortField() == "" {
		query = query.Order("iac_env.created_at DESC")
	} else {
		query = form.Order(query)
	}

	p := page.New(form.CurrentPage(), form.PageSize(), query)
	details := make([]*models.EnvDetail, 0)
	if err := p.Scan(&details); err != nil {
		return nil, e.New(e.DBError, err)
	}

	if details != nil {
		for _, env := range details {
			env.MergeTaskStatus()
			env = PopulateLastTask(c.DB(), env)
		}
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     details,
	}, nil
}

// PopulateLastTask 导出 last task 相关数据
func PopulateLastTask(query *db.Session, env *models.EnvDetail) *models.EnvDetail {
	if env.LastTaskId != "" {
		if lastTask, _ := services.GetTaskById(query, env.LastTaskId); lastTask != nil {
			// 密钥
			if env.KeyId != lastTask.KeyId {
				if key, _ := services.GetKeyById(query, lastTask.KeyId, false); key != nil {
					env.KeyId = lastTask.KeyId
					env.KeyName = key.Name
				}
			}
			// 部署通道
			env.RunnerId = lastTask.RunnerId
			// 分支/标签
			env.Revision = lastTask.Revision
			// Commit id
			env.CommitId = lastTask.CommitId
			// 执行人
			if operator, _ := services.GetUserByIdRaw(query, lastTask.CreatorId); operator != nil {
				env.Operator = operator.Name
				env.OperatorId = lastTask.CreatorId
			}
		}
	}
	return env
}

func checkUserHasApprovalPerm(c *ctx.ServiceContext) error {
	if c.IsSuperAdmin ||
		services.UserHasOrgRole(c.UserId, c.OrgId, consts.OrgRoleAdmin) ||
		services.UserHasProjectRole(c.UserId, c.OrgId, c.ProjectId, consts.ProjectRoleManager) ||
		services.UserHasProjectRole(c.UserId, c.OrgId, c.ProjectId, consts.ProjectRoleApprover) {
		return nil
	}
	return e.New(e.PermDenyApproval, http.StatusForbidden)
}

// UpdateEnv 环境编辑
func UpdateEnv(c *ctx.ServiceContext, form *forms.UpdateEnvForm) (*models.EnvDetail, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update env %s", form.Id))
	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	query := c.DB().Where("iac_env.org_id = ? AND iac_env.project_id = ?", c.OrgId, c.ProjectId)

	env, err := services.GetEnvById(query, form.Id)
	if err != nil && err.Code() != e.EnvNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 项目已归档，不允许编辑
	if env.Archived && !form.Archived {
		return nil, e.New(e.EnvArchived, http.StatusBadRequest)
	}

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}

	if form.HasKey("keyId") {
		attrs["key_id"] = form.KeyId
	}

	if form.HasKey("runnerId") {
		attrs["runner_id"] = form.RunnerId
	}
	if form.HasKey("retryAble") {
		attrs["retryAble"] = form.RetryAble
	}
	if form.HasKey("retryNumber") {
		attrs["retryNumber"] = form.RetryNumber
	}
	if form.HasKey("retryDelay") {
		attrs["retryDelay"] = form.RetryDelay
	}

	if form.HasKey("autoApproval") {
		if form.AutoApproval != env.AutoApproval {
			if err := checkUserHasApprovalPerm(c); err != nil {
				return nil, e.AutoNew(err, e.PermissionDeny)
			}
		}
		attrs["auto_approval"] = form.AutoApproval
	}

	if form.HasKey("stopOnViolation") {
		attrs["StopOnViolation"] = form.StopOnViolation
	}

	if form.HasKey("destroyAt") {
		destroyAt, err := models.Time{}.Parse(form.DestroyAt)
		if err != nil {
			return nil, e.New(e.BadParam, http.StatusBadRequest, err)
		}
		attrs["auto_destroy_at"] = &destroyAt
		attrs["ttl"] = "" // 直接传入了销毁时间，需要同步清空 ttl
	} else if form.HasKey("ttl") {
		ttl, err := services.ParseTTL(form.TTL)
		if err != nil {
			return nil, e.New(e.BadParam, http.StatusBadRequest, err)
		}

		attrs["ttl"] = form.TTL
		if ttl == 0 {
			// ttl 传 0 表示重置销毁时间
			attrs["auto_destroy_at"] = nil
		} else if env.Status != models.EnvStatusInactive {
			// 活跃环境同步修改 destroyAt
			at := models.Time(time.Now().Add(ttl))
			attrs["auto_destroy_at"] = &at
		}
	}

	if form.HasKey("triggers") {
		attrs["triggers"] = pq.StringArray(form.Triggers)
		// triggers有变更时，需要检测webhook的配置
		tpl, err := services.GetTemplateById(c.DB(), env.TplId)
		if err != nil && err.Code() == e.TemplateNotExists {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		} else if err != nil {
			c.Logger().Errorf("error get template, err %s", err)
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		}
		vcs, _ := services.QueryVcsByVcsId(tpl.VcsId, c.DB())
		if err := vcsrv.SetWebhook(vcs, tpl.RepoId, form.Triggers); err != nil {
			c.Logger().Errorf("set webhook err")
		}
	}

	if form.HasKey("archived") {
		if env.Status != models.EnvStatusInactive {
			return nil, e.New(e.EnvCannotArchiveActive,
				fmt.Errorf("env can't be archive while env is %s", env.Status),
				http.StatusBadRequest)
		}
		attrs["archived"] = form.Archived
	}

	env, err = services.UpdateEnv(c.DB(), form.Id, attrs)
	if err != nil && err.Code() == e.EnvAliasDuplicate {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error update env, err %s", err)
		return nil, err
	}

	env.MergeTaskStatus()
	detail := &models.EnvDetail{Env: *env}
	detail = PopulateLastTask(c.DB(), detail)

	return detail, nil
}

// EnvDetail 环境信息详情
func EnvDetail(c *ctx.ServiceContext, form forms.DetailEnvForm) (*models.EnvDetail, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	query := c.DB().Where("iac_env.org_id = ? AND iac_env.project_id = ?", c.OrgId, c.ProjectId)
	query = services.QueryEnvDetail(query)

	envDetail, err := services.GetEnvDetailById(query, form.Id)
	if err != nil && err.Code() == e.EnvNotExists {
		return nil, e.New(e.EnvNotExists, err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get env by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	envDetail.MergeTaskStatus()
	envDetail = PopulateLastTask(c.DB(), envDetail)

	return envDetail, nil
}

// EnvDeploy 创建新部署任务
// 任务类型：plan, apply, destroy
func EnvDeploy(c *ctx.ServiceContext, form *forms.DeployEnvForm) (*models.EnvDetail, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create env task %s", form.Id))
	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	envQuery := services.QueryWithProjectId(services.QueryWithOrgId(tx, c.OrgId), c.ProjectId)
	env, err := services.GetEnvById(envQuery, form.Id)
	if err != nil && err.Code() != e.EnvNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// env 状态检查
	if env.Archived {
		return nil, e.New(e.EnvArchived, http.StatusBadRequest)
	}
	if env.Deploying {
		return nil, e.New(e.EnvDeploying, http.StatusBadRequest)
	}

	// 模板检查
	tplQuery := services.QueryWithOrgId(tx, c.OrgId)
	tpl, err := services.GetTemplateById(tplQuery, env.TplId)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get template, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	if tpl.Status == models.Disable {
		return nil, e.New(e.TemplateDisabled, http.StatusBadRequest)
	}

	if form.HasKey("name") {
		env.Name = form.Name
	}
	if form.HasKey("autoApproval") {
		if form.AutoApproval != env.AutoApproval {
			if err := checkUserHasApprovalPerm(c); err != nil {
				return nil, e.AutoNew(err, e.PermissionDeny)
			}
		}
		env.AutoApproval = form.AutoApproval
	}
	if form.HasKey("stopOnViolation") {
		env.StopOnViolation = form.StopOnViolation
	}

	if form.HasKey("destroyAt") {
		destroyAt, err := models.Time{}.Parse(form.DestroyAt)
		if err != nil {
			return nil, e.New(e.BadParam, http.StatusBadRequest, err)
		}
		env.AutoDestroyAt = &destroyAt
		// 直接传入了销毁时间，需要同步清空 ttl
		env.TTL = ""
	} else if form.HasKey("ttl") {
		ttl, err := services.ParseTTL(form.TTL)
		if err != nil {
			return nil, e.New(e.BadParam, http.StatusBadRequest, err)
		}

		env.TTL = form.TTL

		if ttl == 0 { // ttl 传入 0 表示清空自动销毁时间
			env.AutoDestroyAt = nil
		} else if env.Status != models.EnvStatusInactive {
			// 活跃环境同步修改 destroyAt
			at := models.Time(time.Now().Add(ttl))
			env.AutoDestroyAt = &at
		}
	}

	if form.HasKey("triggers") {
		env.Triggers = form.Triggers
	}

	if form.HasKey("keyId") {
		env.KeyId = form.KeyId
	}
	if form.HasKey("runnerId") {
		env.RunnerId = form.RunnerId
	}
	if form.HasKey("timeout") {
		env.Timeout = form.Timeout
	}

	if form.HasKey("variables") || form.HasKey("deleteVariablesId") {
		// 变量列表增删
		if err = services.OperationVariables(tx, c.OrgId, c.ProjectId, env.TplId, env.Id, form.Variables, form.DeleteVariablesId); err != nil {
			return nil, e.New(err.Code(), err, http.StatusInternalServerError)
		}
	}

	if form.HasKey("tfVarsFile") {
		env.TfVarsFile = form.TfVarsFile
	}
	if form.HasKey("playVarsFile") {
		env.PlayVarsFile = form.PlayVarsFile
	}
	if form.HasKey("playbook") {
		env.Playbook = form.Playbook
	}
	if form.HasKey("revision") {
		env.Revision = form.Revision
	}
	if form.HasKey("retryAble") {
		env.RetryAble = form.RetryAble
	}
	if form.HasKey("retryNumber") {
		env.RetryNumber = form.RetryNumber
	}
	if form.HasKey("retryDelay") {
		env.RetryDelay = form.RetryDelay
	}

	if form.TaskType == "" {
		return nil, e.New(e.BadParam, http.StatusBadRequest)
	}

	targets := make([]string, 0)
	if len(strings.TrimSpace(form.Targets)) > 0 {
		targets = strings.Split(strings.TrimSpace(form.Targets), ",")
	}
	if form.HasKey("varGroupIds") || form.HasKey("delVarGroupIds") {
		// 创建变量组与实例的关系
		if err := services.BatchUpdateRelationship(tx, form.VarGroupIds, form.DelVarGroupIds, consts.ScopeEnv, env.Id.String()); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	// 计算变量列表
	vars, er := services.GetValidVarsAndVgVars(tx, env.OrgId, env.ProjectId, env.TplId, env.Id)
	if er != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// 创建任务
	task, err := services.CreateTask(tx, tpl, env, models.Task{
		Name:            models.Task{}.GetTaskNameByType(form.TaskType),
		Targets:         targets,
		CreatorId:       c.UserId,
		KeyId:           env.KeyId,
		Variables:       vars,
		AutoApprove:     env.AutoApproval,
		Revision:        env.Revision,
		StopOnViolation: env.StopOnViolation,
		BaseTask: models.BaseTask{
			Type:        form.TaskType,
			StepTimeout: form.Timeout,
			RunnerId:    env.RunnerId,
		},
	})

	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating task, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	// Save() 调用会全量将结构体中的字段进行保存，即使字段为 zero value
	if _, err := tx.Save(env); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error save env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit env, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	env.MergeTaskStatus()
	envDetail := &models.EnvDetail{
		Env:    *env,
		TaskId: task.Id,
	}
	envDetail = PopulateLastTask(c.DB(), envDetail)
	vcs, _ := services.QueryVcsByVcsId(tpl.VcsId, c.DB())
	if err := vcsrv.SetWebhook(vcs, tpl.RepoId, form.Triggers); err != nil {
		c.Logger().Errorf("set webhook err :%v", err)
	}
	return envDetail, nil
}

// SearchEnvResources 查询环境资源列表
func SearchEnvResources(c *ctx.ServiceContext, form *forms.SearchEnvResourceForm) (interface{}, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" || form.Id == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	env, err := services.GetEnvById(c.DB(), form.Id)
	if err != nil && err.Code() != e.EnvNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 无资源变更
	if env.LastResTaskId == "" {
		return getEmptyListResult(form)
	}

	return SearchTaskResources(c, &forms.SearchTaskResourceForm{
		PageForm: form.PageForm,
		Id:       env.LastResTaskId,
		Q:        form.Q,
	})
}

// EnvOutput 环境的 Terraform output
// output 与 resource 返回源保持一致
func EnvOutput(c *ctx.ServiceContext, form forms.DetailEnvForm) (interface{}, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" || form.Id == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	env, err := services.GetEnvById(c.DB(), form.Id)
	if err != nil && err.Code() != e.EnvNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 无资源变更
	if env.LastResTaskId == "" {
		return nil, nil
	}

	return TaskOutput(c, forms.DetailTaskForm{
		BaseForm: form.BaseForm,
		Id:       env.LastResTaskId,
	})
}

// EnvVariables 环境部署对应的环境变量为 last task 固化的变量内容
func EnvVariables(c *ctx.ServiceContext, form forms.SearchEnvVariableForm) (interface{}, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" || form.Id == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	env, err := services.GetEnvById(c.DB(), form.Id)
	if err != nil && err.Code() != e.EnvNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	if env.LastTaskId == "" {
		return nil, nil
	}

	task, err := services.GetTaskById(c.DB(), env.LastTaskId)
	if err != nil && err.Code() != e.TaskNotExists {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	} else if err != nil {
		c.Logger().Errorf("error get env last res task, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 隐藏敏感字段
	for index, v := range task.Variables {
		if v.Sensitive {
			task.Variables[index].Value = ""
		}
	}
	sort.Sort(task.Variables)
	return task.Variables, nil
}

// ResourceDetail 查询部署成功后资源的详细信息
func ResourceDetail(c *ctx.ServiceContext, form *forms.ResourceDetailForm) (*models.ResAttrs, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" || form.Id == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	resource, err := services.GetResourceById(c.DB(), form.ResourceId)
	if err != nil {
		c.Logger().Errorf("error get resource, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	if resource.EnvId != form.Id || resource.OrgId != c.OrgId || resource.ProjectId != c.ProjectId {
		c.Logger().Errorf("Environment ID and resource ID do not match")
		return nil, e.New(e.DBError, err, http.StatusForbidden)
	}
	resultAttrs := resource.Attrs
	if len(resource.SensitiveKeys) > 0 {
		set := map[string]interface{}{}
		for _, value := range resource.SensitiveKeys {
			set[value] = nil
		}
		for k, _ := range resultAttrs {
			// 如果state 中value 存在与sensitive 设置，展示时不展示详情
			if _, ok := set[k]; ok {
				resultAttrs[k] = "(sensitive value)"
			}
		}
	}
	return &resultAttrs, nil
}

// SearchEnvResourcesGraph 查询环境资源列表
func SearchEnvResourcesGraph(c *ctx.ServiceContext, form *forms.SearchEnvResourceGraphForm) (interface{}, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" || form.Id == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	env, err := services.GetEnvById(c.DB(), form.Id)
	if err != nil && err.Code() != e.EnvNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 无资源变更
	if env.LastResTaskId == "" {
		return nil, nil
	}

	return SearchTaskResourcesGraph(c, &forms.SearchTaskResourceGraphForm{
		Id:        env.LastResTaskId,
		Dimension: form.Dimension,
	})
}

// ResourceGraphDetail 查询部署成功后资源的详细信息
func ResourceGraphDetail(c *ctx.ServiceContext, form *forms.ResourceGraphDetailForm) (*models.Resource, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" || form.Id == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	resource, err := services.GetResourceById(c.DB(), form.ResourceId)
	if err != nil {
		c.Logger().Errorf("error get resource, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	if resource.EnvId != form.Id || resource.OrgId != c.OrgId || resource.ProjectId != c.ProjectId {
		c.Logger().Errorf("Environment ID and resource ID do not match")
		return nil, e.New(e.DBError, err, http.StatusForbidden)
	}
	resultAttrs := resource.Attrs
	if len(resource.SensitiveKeys) > 0 {
		set := map[string]interface{}{}
		for _, value := range resource.SensitiveKeys {
			set[value] = nil
		}
		for k, _ := range resultAttrs {
			// 如果state 中value 存在与sensitive 设置，展示时不展示详情
			if _, ok := set[k]; ok {
				resultAttrs[k] = "(sensitive value)"
			}
		}
	}
	resource.Attrs = resultAttrs
	return resource, nil
}
