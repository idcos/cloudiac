package apps

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// CreateEnv 创建环境
func CreateEnv(c *ctx.ServiceCtx, form *forms.CreateEnvForm) (*models.Env, e.Error) {
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
	if form.TfVarsFile == "" {
		form.TfVarsFile = tpl.TfVarsFile
	}
	if form.PlayVarsFile == "" {
		form.PlayVarsFile = tpl.PlayVarsFile
	}
	if form.Playbook == "" {
		form.Playbook = tpl.Playbook
	}
	if form.Timeout == 0 {
		form.Timeout = common.TaskStepTimeoutDuration
	}

	var (
		DestroyAt *time.Time
	)

	if form.DestroyAt != "" {
		at, err := time.Parse("2006-01-02 15:04", form.DestroyAt)
		if err != nil {
			return nil, e.New(e.BadParam, http.StatusBadRequest)
		}
		DestroyAt = &at
	} else if form.TTL != "" {
		_, err := time.ParseDuration(form.TTL)
		if err != nil {
			return nil, e.New(e.BadParam, http.StatusBadRequest)
		}
	}

	tx := c.Tx()
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
		Timeout:  form.Timeout,

		// 模板参数
		TfVarsFile:   form.TfVarsFile,
		PlayVarsFile: form.PlayVarsFile,
		Playbook:     form.Playbook,
		Revision:     form.Revision,
		KeyId:        form.KeyId,

		TTL:           form.TTL,
		AutoDestroyAt: DestroyAt,
		AutoApproval:  form.AutoApproval,

		Triggers: form.Triggers,
	})
	if err != nil && err.Code() == e.EnvAlreadyExists {
		_ = tx.Rollback()
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating env, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	// 创建新导入的变量
	if err = services.OperationVariables(tx, c.OrgId, c.ProjectId, env.TplId, env.Id, form.Variables, nil); err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}
	// 获取计算后的变量列表
	vars, err, _ := services.GetValidVariables(tx, consts.ScopeEnv, c.OrgId, c.ProjectId, env.TplId, env.Id)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}
	// 保存固化的变量
	env.Variables = getVariables(vars)

	targets := make([]string, 0)
	if len(strings.TrimSpace(form.Targets)) > 0 {
		targets = strings.Split(strings.TrimSpace(form.Targets), ",")
	}

	// 创建任务
	task, err := services.CreateTask(tx, tpl, env, models.Task{
		Name:        models.Task{}.GetTaskNameByType(form.TaskType),
		Type:        form.TaskType,
		Flow:        models.TaskFlow{},
		Targets:     targets,
		CreatorId:   c.UserId,
		KeyId:       env.KeyId,
		RunnerId:    env.RunnerId,
		Variables:   getVariableBody(env.Variables),
		StepTimeout: form.Timeout,
		AutoApprove: env.AutoApproval,
	})
	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating task, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

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

	return env, nil
}

func getVariables(vars map[string]models.Variable) models.EnvVariables {
	var vb []models.Variable
	for _, v := range vars {
		vb = append(vb, v)
	}
	return vb
}

func getVariableBody(vars models.EnvVariables) []models.VariableBody {
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
		query = query.WhereLike("iac_env.name", form.Q)
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

	if form.HasKey("keyId") {
		attrs["key_id"] = form.KeyId
	}

	if form.HasKey("runnerId") {
		attrs["runner_id"] = form.RunnerId
	}

	if form.HasKey("autoApproval") {
		attrs["auto_approval"] = form.AutoApproval
	}

	if form.HasKey("destroyAt") {
		at, err := time.Parse("2006-01-02 15:04", form.DestroyAt)
		if err != nil {
			return nil, e.New(e.BadParam, http.StatusBadRequest)
		}
		attrs["auto_destroy_at"] = &at
	} else if form.HasKey("ttl") {
		ttl, err := time.ParseDuration(form.TTL)
		if err != nil {
			return nil, e.New(e.BadParam, http.StatusBadRequest)
		}

		attrs["ttl"] = form.TTL
		if env.LastTaskId != "" { // 己部署过的环境同步修改 destroyAt
			at := time.Now().Add(ttl)
			attrs["auto_destroy_at"] = &at
		}
	}

	if form.HasKey("triggers") {
		env.Triggers = form.Triggers
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

// EnvDeploy 创建新部署任务
// 任务类型：apply, destroy
func EnvDeploy(c *ctx.ServiceCtx, form *forms.DeployEnvForm) (*models.Env, e.Error) {
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

	if form.TfVarsFile == "" {
		form.TfVarsFile = tpl.TfVarsFile
	}
	if form.PlayVarsFile == "" {
		form.PlayVarsFile = tpl.PlayVarsFile
	}
	if form.Playbook == "" {
		form.Playbook = tpl.Playbook
	}

	if form.HasKey("name") {
		env.Name = form.Name
	}
	if form.HasKey("autoApproval") {
		env.AutoApproval = form.AutoApproval
	}

	if form.HasKey("destroyAt") {
		at, err := time.Parse("2006-01-02 15:04", form.DestroyAt)
		if err != nil {
			return nil, e.New(e.BadParam, http.StatusBadRequest)
		}
		env.AutoDestroyAt = &at
	} else if form.HasKey("ttl") {
		ttl, err := time.ParseDuration(form.TTL)
		if err != nil {
			return nil, e.New(e.BadParam, http.StatusBadRequest)
		}

		env.TTL = form.TTL
		if env.LastTaskId != "" { // 己部署过的环境同步修改 destroyAt
			at := time.Now().Add(ttl)
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
		// 计算变量列表
		vars, err, _ := services.GetValidVariables(tx, consts.ScopeEnv, c.OrgId, c.ProjectId, env.TplId, env.Id)
		if err != nil {
			return nil, e.New(err.Code(), err, http.StatusInternalServerError)
		}
		// 保存固化变量
		env.Variables = getVariables(vars)
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
	if form.TaskType == "" {
		return nil, e.New(e.BadParam, http.StatusBadRequest)
	}

	targets := make([]string, 0)
	if len(strings.TrimSpace(form.Targets)) > 0 {
		targets = strings.Split(strings.TrimSpace(form.Targets), ",")
	}

	// 创建任务
	task, err := services.CreateTask(tx, tpl, env, models.Task{
		Name:        models.Task{}.GetTaskNameByType(form.TaskType),
		Type:        form.TaskType,
		Flow:        models.TaskFlow{},
		Targets:     targets,
		CreatorId:   c.UserId,
		KeyId:       env.KeyId,
		RunnerId:    env.RunnerId,
		Variables:   getVariableBody(env.Variables),
		StepTimeout: form.Timeout,
		AutoApprove: env.AutoApproval,
	})

	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating task, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	env.LastTaskId = task.Id
	if _, err := tx.Save(env); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error save env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

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

	return env, nil
}

// SearchEnvResources 查询环境资源列表
func SearchEnvResources(c *ctx.ServiceCtx, form *forms.SearchEnvResourceForm) (interface{}, e.Error) {
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

	query := c.DB().Model(models.Resource{}).Where("org_id = ? AND project_id = ? AND env_id = ? AND task_id = ?",
		c.OrgId, c.ProjectId, form.Id, env.LastTaskId)

	if form.HasKey("q") {
		// 支持对 provider / type / name 进行模糊查询
		query = query.Where("provider LIKE ? OR type LIKE ? OR name LIKE ?",
			fmt.Sprintf("%%%s%%", form.Q),
			fmt.Sprintf("%%%s%%", form.Q),
			fmt.Sprintf("%%%s%%", form.Q))
	}

	if form.SortField() == "" {
		query = query.Order("provider, type, name")
	}

	return getPage(query, form, &models.Resource{})
}

// SearchEnvVariables 查询环境变量列表
func SearchEnvVariables(c *ctx.ServiceCtx, form *forms.SearchEnvVariableForm) (interface{}, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" || form.Id == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	query := c.DB().Where("org_id = ? AND project_id = ? AND id = ?", c.OrgId, c.ProjectId, form.Id)
	env, err := services.GetEnvById(query, form.Id)
	if err != nil && err.Code() == e.EnvNotExists {
		return nil, e.New(e.EnvNotExists, err, http.NotFound)
	} else if err != nil {
		c.Logger().Errorf("error while get env by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	rs := make([]VariableResp, 0)
	for _, variable := range env.Variables {
		vr := VariableResp{
			Variable:   variable,
			Overwrites: nil,
		}
		isExists, overwrites := services.GetVariableParent(c.DB(), variable.Name, variable.Scope, variable.Type, common.EnvScopeEnv)
		if isExists {
			vr.Overwrites = &overwrites
		}

		rs = append(rs, vr)
	}
	sort.Sort(newVariable(rs))

	return env.Variables, nil
}
