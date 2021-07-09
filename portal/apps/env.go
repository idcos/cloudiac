package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/portal/services/vcsrv"
	"cloudiac/utils"
	"fmt"
	"net/http"
)

// CreateEnv 创建环境
func CreateEnv(c *ctx.ServiceCtx, form *forms.CreateEnvForm) (*models.Env, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create env %s", form.Name))

	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	// 模板
	query := c.DB().Where("status = %s", models.Enable)
	tpl, err := services.GetTemplateById(query, form.TplId)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get template, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	if form.TfVarsFile == "" {
		form.TfVarsFile = tpl.TfVarsFile
	}
	if form.PlayVarsFile == "" {
		form.PlayVarsFile = tpl.PlayVarsFile
	}
	if form.Playbook == "" {
		form.Playbook = tpl.Playbook
	}

	// 获取最新 commit id
	vcs, err := services.QueryVcsByVcsId(tpl.VcsId, c.DB())
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}
	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.New(e.GitLabError, er, http.StatusInternalServerError)
	}
	repo, er := vcsService.GetRepo(tpl.RepoId)
	if er != nil {
		return nil, e.New(e.GitLabError, er, http.StatusInternalServerError)
	}
	revision := tpl.RepoRevision
	if form.Revision != "" {
		revision = form.Revision
	}
	commitId, er := repo.BranchCommitId(revision)
	if er != nil {
		return nil, e.New(e.GitLabError, er, http.StatusInternalServerError)
	}

	// 变量
	// TODO: 检查、保存、合并环境变量
	vars := form.Variables

	tx := c.DB()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	env, err := services.CreateEnv(tx, models.Env{
		OrgId:     c.OrgId,
		ProjectId: c.ProjectId,
		CreatorId: c.UserId,
		TplId:     form.TplId,

		Name:     form.Name,
		RunnerId: form.RunnerId,
		Status:   models.EnvStatusInactive,
		OneTime:  form.OneTime,

		// 模板参数
		TfVarsFile:   form.TfVarsFile,
		PlayVarsFile: form.PlayVarsFile,
		Playbook:     form.Playbook,

		// 变量参数
		Variables: vars,

		// TODO: triggers 触发器设置
		AutoApproval: form.AutoApproval,
		// TODO: 自动销毁设置

	})
	if err != nil && err.Code() == e.EnvAlreadyExists {
		_ = tx.Rollback()
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating env, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	// 创建任务
	flow, er := models.DefaultTaskFlow(form.TaskType)
	if er != nil {
		return nil, e.New(e.BadParam, er, http.StatusBadRequest)
	}
	task, err := services.CreateTask(tx, env, models.Task{
		CreatorId: c.UserId,
		RunnerId:  env.RunnerId,
		CommitId:  commitId,
		Type:      form.TaskType,
		Name:      models.Task{}.GetTaskNameByType(form.TaskType),
		Flow:      flow,
		Variables: form.Variables,
	})
	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating task, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	env.CurrentTaskId = task.Id
	if _, er = tx.Save(env); er != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error save env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit env, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return env, nil
}

// SearchEnv 环境查询
func SearchEnv(c *ctx.ServiceCtx, form *forms.SearchEnvForm) (interface{}, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	query := c.DB().Where("iac_env.org_id = ? AND iac_env.project_id = ?", c.OrgId, c.ProjectId)
	query = services.QueryEnvDetail(query)

	if form.Status != "" {
		if utils.InArrayStr(models.EnvStatus, form.Status) {
			return nil, e.New(e.BadParam, http.StatusBadRequest)
		}
		query = query.Where("iac_env.status = ?", form.Status)
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
		query = query.Where("iac_env.archived == 1")
	case "false":
		// 未归档
		query = query.Where("iac_env.archived == 0")
	default:
		return nil, e.New(e.BadParam, http.StatusBadRequest)
	}

	if form.Q != "" {
		qs := "%" + form.Q + "%"
		query = query.WhereLike("iac_env.name", qs)
	}

	// 默认按创建时间逆序排序
	if form.SortField() == "" {
		query = query.Order("iac_env.created_at DESC")
	}

	rs, err := getPage(query, form, &models.EnvDetail{})
	if err != nil {
		c.Logger().Errorf("error get page, err %s", err)
	}
	return rs, err
}

// UpdateEnv 环境编辑
func UpdateEnv(c *ctx.ServiceCtx, form *forms.UpdateEnvForm) (*models.Env, e.Error) {
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
	if env.Archived && form.Archived == false {
		return nil, e.New(e.EnvArchived, http.StatusBadRequest)
	}

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}

	if form.HasKey("runnerId") {
		attrs["runner_id"] = form.RunnerId
	}

	if form.HasKey("autoApproval") {
		attrs["auto_approval"] = form.AutoApproval
	}

	if form.HasKey("autoDestroyAt") {
		// TODO: 修改生命周期
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
	return env, nil
}

// EnvDetail 环境信息详情
func EnvDetail(c *ctx.ServiceCtx, form forms.DetailEnvForm) (*models.EnvDetail, e.Error) {
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

	return envDetail, nil
}

// EnvDeploy 创建任务
func EnvDeploy(c *ctx.ServiceCtx, form *forms.DeployEnvForm) (*models.Env, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create env task %s", form.Id))
	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	query := c.DB().Where("iac_env.org_id = ? AND iac_env.project_id = ?", c.OrgId, c.ProjectId)
	query = query.Where("iac_env.archived = 0")
	env, err := services.GetEnvById(query, form.Id)
	if err != nil && err.Code() != e.EnvNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 模板
	query = c.DB().Where("status = %s", models.Enable)
	tpl, err := services.GetTemplateById(query, env.TplId)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get template, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	if form.TfVarsFile == "" {
		form.TfVarsFile = tpl.TfVarsFile
	}
	if form.PlayVarsFile == "" {
		form.PlayVarsFile = tpl.PlayVarsFile
	}
	if form.Playbook == "" {
		form.Playbook = tpl.Playbook
	}

	// 获取最新 commit id
	vcs, err := services.QueryVcsByVcsId(tpl.VcsId, c.DB())
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}
	vcsService, er := vcsrv.GetVcsInstance(vcs)
	if er != nil {
		return nil, e.New(e.GitLabError, er, http.StatusInternalServerError)
	}
	repo, er := vcsService.GetRepo(tpl.RepoId)
	if er != nil {
		return nil, e.New(e.GitLabError, er, http.StatusInternalServerError)
	}
	revision := tpl.RepoRevision
	if form.Revision != "" {
		revision = form.Revision
	}
	commitId, er := repo.BranchCommitId(revision)
	if er != nil {
		return nil, e.New(e.GitLabError, er, http.StatusInternalServerError)
	}

	// 变量
	// TODO: 检查、保存、合并环境变量
	//vars := form.Variables

	tx := c.DB()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	if form.HasKey("name") {
		env.Name = form.Name
	}
	if form.HasKey("autoApproval") {
		env.AutoApproval = form.AutoApproval
	}
	if form.HasKey("autoDestroyAt") {
		// TODO: 自动销毁设置
	}
	if form.HasKey("runnerId") {
		env.RunnerId = form.RunnerId
	}
	if form.HasKey("timeout") {
		env.Timeout = form.Timeout
	}
	if form.HasKey("variables") {
		env.Variables = form.Variables
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
	if form.HasKey("taskType") {
		if form.TaskType == "" {
			form.TaskType = models.TaskTypeApply
		}
	}

	// 创建任务
	flow, er := models.DefaultTaskFlow(form.TaskType)
	if er != nil {
		return nil, e.New(e.BadParam, er, http.StatusBadRequest)
	}
	task, err := services.CreateTask(tx, env, models.Task{
		CreatorId: c.UserId,
		RunnerId:  env.RunnerId,
		CommitId:  commitId,
		Type:      form.TaskType,
		Name:      models.Task{}.GetTaskNameByType(form.TaskType),
		Flow:      flow,
		Variables: form.Variables,
	})
	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating task, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	// 如果环境为不活跃或者失败状态，立即发起任务部署，设置这个任务为当前任务
	if env.Status == models.EnvStatusInactive || env.Status == models.EnvStatusFailed {
		env.CurrentTaskId = task.Id
	}
	if _, er = tx.Save(env); er != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error save env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error commit env, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return env, nil
}
