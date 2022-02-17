// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

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
	"cloudiac/utils/logs"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/lib/pq"
)

// 最小时间单位为分钟
// 每隔1 分钟执行一次 */1 * * * ?
// 每天 23点 执行一次 0 23 * * ?
// 每个月1号23 点执行一次 0 23 1 * ?
// 每天的0点、13点、18点、21点都执行一次：0 0,13,18,21 * * ?
var SpecParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

func ParseCronpress(cronDriftExpress string) (*time.Time, e.Error) {
	expr, err := SpecParser.Parse(cronDriftExpress)
	if err != nil {
		return nil, e.New(e.BadParam, http.StatusBadRequest, err)
	}
	// 根据当前时间算出下次该环境下次执行偏移检测任务时间
	nextTime := expr.Next(time.Now())
	return &nextTime, nil
}

// 获取环境偏移检测任务类型并且检查其参数
func GetCronTaskTypeAndCheckParam(cronExpress string, autoRepairDrift, openCronDrift bool) (string, e.Error) {
	if openCronDrift {
		if cronExpress == "" {
			return "", e.New(e.BadParam, http.StatusBadRequest, "Please set cronExpress when openCronDrift is set")
		}
	}
	if autoRepairDrift {
		if !openCronDrift || cronExpress == "" {
			return "", e.New(e.BadParam, http.StatusBadRequest, "Please set openCronDrift to true when autoRepairDrift is set")
		}
	}
	if openCronDrift {
		if autoRepairDrift == false {
			return models.TaskTypePlan, nil
		} else {
			return models.TaskTypeApply, nil
		}
	}
	// 未开启漂移检测任务
	return "", nil
}

type CronDriftParam struct {
	CronDriftExpress  *string    `json:"cronDriftExpress"`  // 偏移检测表达式
	AutoRepairDrift   *bool      `json:"autoRepairDrift"`   // 是否进行自动纠偏
	OpenCronDrift     *bool      `json:"openCronDrift"`     // 是否开启偏移检测
	NextDriftTaskTime *time.Time `json:"nextDriftTaskTime"` // 下次执行偏移检测任务的时间
}

func GetCronDriftParam(form forms.CronDriftForm) (*CronDriftParam, e.Error) {
	cronDriftParam := &CronDriftParam{}
	if form.HasKey("cronDriftExpress") || form.HasKey("autoRepairDrift") || form.HasKey("openCronDrift") {
		cronTaskType, err := GetCronTaskTypeAndCheckParam(form.CronDriftExpress, form.AutoRepairDrift, form.OpenCronDrift)
		if err != nil {
			return nil, err
		}
		if form.HasKey("autoRepairDrift") {
			cronDriftParam.AutoRepairDrift = &form.AutoRepairDrift
		}
		if form.HasKey("openCronDrift") {
			cronDriftParam.OpenCronDrift = &form.OpenCronDrift
		}
		if cronTaskType != "" {
			// 如果任务类型不为空，说明配置了漂移检测任务
			cronDriftParam.CronDriftExpress = &form.CronDriftExpress
			nextTime, err := ParseCronpress(form.CronDriftExpress)
			if err != nil {
				return nil, err
			}
			cronDriftParam.NextDriftTaskTime = nextTime
		} else {
			// 更新配置取消漂移检测任务，将下次重试时间重置为nil
			cronDriftParam.NextDriftTaskTime = nil
		}
	}
	return cronDriftParam, nil
}

// CreateEnv 创建环境
func CreateEnv(c *ctx.ServiceContext, form *forms.CreateEnvForm) (*models.EnvDetail, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create env %s", form.Name))

	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	// 检查自动纠漂移、推送到分支时重新部署时，是否了配置自动审批
	if !services.CheckoutAutoApproval(form.AutoApproval, form.AutoRepairDrift, form.Triggers) {
		return nil, e.New(e.EnvCheckAutoApproval, http.StatusBadRequest)
	}

	if form.Playbook != "" && form.KeyId == "" {
		return nil, e.New(e.TemplateKeyIdNotSet)
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
	if !form.HasKey("keyId") {
		form.KeyId = tpl.KeyId
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

	envModel := models.Env{
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

		ExtraData:        models.JSON(form.ExtraData),
		Callback:         form.Callback,
		AutoRepairDrift:  form.AutoRepairDrift,
		CronDriftExpress: form.CronDriftExpress,
		OpenCronDrift:    form.OpenCronDrift,
		PolicyEnable:     form.PolicyEnable,
	}
	// 检查偏移检测参数
	cronTaskType, err := GetCronTaskTypeAndCheckParam(form.CronDriftExpress, form.AutoRepairDrift, form.OpenCronDrift)
	if err != nil {
		return nil, err
	}
	// 如果定时任务存在，保存参数到表内容
	if cronTaskType != "" {
		nextTime, err := ParseCronpress(form.CronDriftExpress)
		if err != nil {
			return nil, err
		}
		envModel.NextDriftTaskTime = nextTime
	}
	env, err := services.CreateEnv(tx, envModel)
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

	// 创建变量
	updateVarsForm := forms.UpdateObjectVarsForm{
		Scope:     consts.ScopeEnv,
		ObjectId:  env.Id,
		Variables: form.Variables,
	}
	if _, er := updateObjectVars(c, tx, &updateVarsForm); er != nil {
		_ = tx.Rollback()
		return nil, e.AutoNew(er, e.InternalError)
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
	// 绑定策略组
	if len(form.PolicyGroup) > 0 {
		policyForm := &forms.UpdatePolicyRelForm{
			Id:             env.Id,
			Scope:          consts.ScopeEnv,
			PolicyGroupIds: form.PolicyGroup,
		}
		if _, err = services.UpdatePolicyRel(tx, policyForm); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	// 来源：手动触发、外部调用
	taskSource := consts.TaskSourceManual
	taskSourceSys := ""
	if form.Source != "" || form.Callback != "" {
		taskSource = consts.TaskSourceApi
		taskSourceSys = form.Source
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
		Source:    taskSource,
		SourceSys: taskSourceSys,
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
	// 获取token
	token, err := GetWebhookToken(c)
	if err != nil {
		return nil, err
	}

	if err := vcsrv.SetWebhook(vcs, tpl.RepoId, token.Key, form.Triggers); err != nil {
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
		query = query.Joins("left join iac_template on iac_env.tpl_id = iac_template.id")
		query = query.Where("iac_env.name LIKE ? OR iac_template.name LIKE ?",
			fmt.Sprintf("%%%s%%", form.Q),
			fmt.Sprintf("%%%s%%", form.Q),
		)
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

	for _, env := range details {
		env.MergeTaskStatus()
		// FIXME: 这里会在 for 循环中查询 db，需要优化
		PopulateLastTask(c.DB(), env)
		env.PolicyStatus = models.PolicyStatusConversion(env.PolicyStatus, env.PolicyEnable)
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
			// 返回环境详情时会调用该函数，且返回的数据会用于创建新的部署任务时在表单中回显，
			// 而对于 vcs webhook 自动触发的任务其 revision 可能与环境的不同，这就会导致重新部署环境时使用的不是环境当前配置的分支。
			// 为了避免这种情况，这里我们不将环境的 Revision 设置为最后一次任务的 revision
			// env.Revision = lastTask.Revision // 分支/标签

			// 部署通道
			env.RunnerId = lastTask.RunnerId
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

func updateEnvCheck(orgId, projectId models.Id, form *forms.UpdateEnvForm) e.Error {
	if orgId == "" || projectId == "" {
		return e.New(e.BadRequest, http.StatusBadRequest)
	}

	// 检查自动纠漂移、推送到分支时重新部署时，是否了配置自动审批
	if !services.CheckoutAutoApproval(form.AutoApproval, form.AutoRepairDrift, form.Triggers) {
		return e.New(e.EnvCheckAutoApproval, http.StatusBadRequest)
	}

	return nil
}

func getEnvForUpdate(tx *db.Session, c *ctx.ServiceContext, form *forms.UpdateEnvForm) (*models.Env, e.Error) {
	query := tx.Where("iac_env.org_id = ? AND iac_env.project_id = ?", c.OrgId, c.ProjectId)

	env, err := services.GetEnvById(query, form.Id)
	if err != nil && err.Code() != e.EnvNotExists {
		_ = tx.Rollback()
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error get env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 项目已归档，不允许编辑
	if env.Archived && !form.Archived {
		return nil, e.New(e.EnvArchived, http.StatusBadRequest)
	}

	return env, nil
}

func setUpdateEnvByForm(attrs models.Attrs, form *forms.UpdateEnvForm) {
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
	if form.HasKey("stopOnViolation") {
		attrs["StopOnViolation"] = form.StopOnViolation
	}
	if form.HasKey("policyEnable") {
		attrs["policyEnable"] = form.PolicyEnable
	}
}

func setAndCheckUpdateEnvAutoApproval(c *ctx.ServiceContext, tx *db.Session, attrs models.Attrs, env *models.Env, form *forms.UpdateEnvForm) e.Error {
	if form.HasKey("autoApproval") {
		if form.AutoApproval != env.AutoApproval {
			if err := checkUserHasApprovalPerm(c); err != nil {
				_ = tx.Rollback()
				return e.AutoNew(err, e.PermissionDeny)
			}
		}
		attrs["auto_approval"] = form.AutoApproval
	}
	return nil
}

func setAndCheckUpdateEnvDestroy(tx *db.Session, attrs models.Attrs, env *models.Env, form *forms.UpdateEnvForm) e.Error {
	if form.HasKey("destroyAt") {
		destroyAt, err := models.Time{}.Parse(form.DestroyAt)
		if err != nil {
			_ = tx.Rollback()
			return e.New(e.BadParam, http.StatusBadRequest, err)
		}
		attrs["auto_destroy_at"] = &destroyAt
		attrs["ttl"] = "" // 直接传入了销毁时间，需要同步清空 ttl
	} else if form.HasKey("ttl") {
		ttl, err := services.ParseTTL(form.TTL)
		if err != nil {
			_ = tx.Rollback()
			return e.New(e.BadParam, http.StatusBadRequest, err)
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
	return nil
}

func setAndCheckUpdateEnvTriggers(c *ctx.ServiceContext, tx *db.Session, attrs models.Attrs, env *models.Env, form *forms.UpdateEnvForm) e.Error {
	if form.HasKey("triggers") {
		attrs["triggers"] = pq.StringArray(form.Triggers)
		// triggers有变更时，需要检测webhook的配置
		tpl, err := services.GetTemplateById(c.DB(), env.TplId)
		if err != nil && err.Code() == e.TemplateNotExists {
			_ = tx.Rollback()
			return e.New(err.Code(), err, http.StatusBadRequest)
		} else if err != nil {
			c.Logger().Errorf("error get template, err %s", err)
			_ = tx.Rollback()
			return e.New(e.DBError, err, http.StatusInternalServerError)
		}
		vcs, _ := services.QueryVcsByVcsId(tpl.VcsId, c.DB())
		// 获取token
		token, err := GetWebhookToken(c)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		if err := vcsrv.SetWebhook(vcs, tpl.RepoId, token.Key, form.Triggers); err != nil {
			c.Logger().Errorf("set webhook err：%v", err)
		}
	}
	return nil
}

func setAndCheckUpdateEnvByForm(c *ctx.ServiceContext, tx *db.Session, attrs models.Attrs, env *models.Env, form *forms.UpdateEnvForm) e.Error {

	if err := setAndCheckUpdateEnvAutoApproval(c, tx, attrs, env, form); err != nil {
		return err
	}

	if err := setAndCheckUpdateEnvDestroy(tx, attrs, env, form); err != nil {
		return err
	}

	if err := setAndCheckUpdateEnvTriggers(c, tx, attrs, env, form); err != nil {
		return err
	}

	if form.HasKey("archived") {
		if env.Status != models.EnvStatusInactive {
			_ = tx.Rollback()
			return e.New(e.EnvCannotArchiveActive,
				fmt.Errorf("env can't be archive while env is %s", env.Status),
				http.StatusBadRequest)
		}
		attrs["archived"] = form.Archived
	}

	return nil
}

// UpdateEnv 环境编辑
func UpdateEnv(c *ctx.ServiceContext, form *forms.UpdateEnvForm) (*models.EnvDetail, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update env %s", form.Id))

	if err := updateEnvCheck(c.OrgId, c.ProjectId, form); err != nil {
		return nil, err
	}

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	env, err := getEnvForUpdate(tx, c, form)
	if err != nil {
		return nil, err
	}

	attrs := models.Attrs{}

	cronDriftParam, err := GetCronDriftParam(forms.CronDriftForm{
		BaseForm:         form.BaseForm,
		CronDriftExpress: form.CronDriftExpress,
		AutoRepairDrift:  form.AutoRepairDrift,
		OpenCronDrift:    form.OpenCronDrift,
	})

	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	// 先更新合规绑定，若失败则不进行后续更新操作
	if form.HasKey("policyGroup") && len(form.PolicyGroup) > 0 {
		policyForm := &forms.UpdatePolicyRelForm{
			Id:             env.Id,
			Scope:          consts.ScopeEnv,
			PolicyGroupIds: form.PolicyGroup,
		}
		if _, err = services.UpdatePolicyRel(tx, policyForm); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	attrs["autoRepairDrift"] = cronDriftParam.AutoRepairDrift
	attrs["openCronDrift"] = cronDriftParam.OpenCronDrift
	attrs["cronDriftExpress"] = cronDriftParam.CronDriftExpress
	attrs["nextDriftTaskTime"] = cronDriftParam.NextDriftTaskTime

	setUpdateEnvByForm(attrs, form)
	err = setAndCheckUpdateEnvByForm(c, tx, attrs, env, form)
	if err != nil {
		return nil, err
	}

	env, err = services.UpdateEnv(tx, form.Id, attrs)
	if err != nil && err.Code() == e.EnvAliasDuplicate {
		_ = tx.Rollback()
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error update env, err %s", err)
		return nil, err
	}

	env.MergeTaskStatus()
	detail := &models.EnvDetail{Env: *env}
	detail = PopulateLastTask(tx, detail)

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}
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
	resp, err := services.GetPolicyRels(c.DB(), form.Id, consts.ScopeEnv)
	if err != nil {
		return nil, err
	}
	envDetail.PolicyGroup = make([]string, 0)
	for _, v := range resp {
		envDetail.PolicyGroup = append(envDetail.PolicyGroup, v.PolicyGroupId)
	}
	envDetail.PolicyStatus = models.PolicyStatusConversion(envDetail.PolicyStatus, envDetail.PolicyEnable)

	return envDetail, nil
}

// EnvDeploy 创建新部署任务
// 任务类型：plan, apply, destroy
func EnvDeploy(c *ctx.ServiceContext, form *forms.DeployEnvForm) (ret *models.EnvDetail, er e.Error) {
	_ = c.DB().Transaction(func(tx *db.Session) error {
		ret, er = envDeploy(c, tx, form)
		return er
	})
	return ret, er
}

func envPreCheck(orgId, projectId, keyId models.Id, playbook string) e.Error {
	if orgId == "" || projectId == "" {
		return e.New(e.BadRequest, http.StatusBadRequest)
	}

	if playbook != "" && keyId == "" {
		return e.New(e.TemplateKeyIdNotSet)
	}

	return nil
}

func envCheck(tx *db.Session, orgId, projectId, id models.Id, lg logs.Logger) (*models.Env, e.Error) {
	envQuery := services.QueryWithProjectId(services.QueryWithOrgId(tx, orgId), projectId)
	env, err := services.GetEnvById(envQuery, id)
	if err != nil && err.Code() != e.EnvNotExists {
		return nil, e.New(err.Code(), err, http.StatusNotFound)
	} else if err != nil {
		lg.Errorf("error get env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// env 状态检查
	if env.Archived {
		return nil, e.New(e.EnvArchived, http.StatusBadRequest)
	}
	if env.Deploying {
		return nil, e.New(e.EnvDeploying, http.StatusBadRequest)
	}

	return env, nil
}

func envTplCheck(tx *db.Session, orgId, tplId models.Id, lg logs.Logger) (*models.Template, e.Error) {
	tplQuery := services.QueryWithOrgId(tx, orgId)
	tpl, err := services.GetTemplateById(tplQuery, tplId)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		lg.Errorf("error get template, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	if tpl.Status == models.Disable {
		return nil, e.New(e.TemplateDisabled, http.StatusBadRequest)
	}

	return tpl, nil
}

func setEnvByForm(env *models.Env, form *forms.DeployEnvForm) {
	if form.HasKey("name") {
		env.Name = form.Name
	}
	if form.HasKey("stopOnViolation") {
		env.StopOnViolation = form.StopOnViolation
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
	if form.HasKey("policyEnable") {
		env.PolicyEnable = form.PolicyEnable
	}
}

func setAndCheckEnvAutoApproval(c *ctx.ServiceContext, env *models.Env, form *forms.DeployEnvForm) e.Error {
	if form.HasKey("autoApproval") {
		if form.AutoApproval != env.AutoApproval {
			if err := checkUserHasApprovalPerm(c); err != nil {
				return e.AutoNew(err, e.PermissionDeny)
			}
		}
		env.AutoApproval = form.AutoApproval
	}

	return nil
}

func setAndCheckEnvDestroy(env *models.Env, form *forms.DeployEnvForm) e.Error {
	if form.HasKey("destroyAt") {
		destroyAt, err := models.Time{}.Parse(form.DestroyAt)
		if err != nil {
			return e.New(e.BadParam, http.StatusBadRequest, err)
		}
		env.AutoDestroyAt = &destroyAt
		// 直接传入了销毁时间，需要同步清空 ttl
		env.TTL = ""
	} else if form.HasKey("ttl") {
		ttl, err := services.ParseTTL(form.TTL)
		if err != nil {
			return e.New(e.BadParam, http.StatusBadRequest, err)
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

	return nil
}
func setAndCheckEnvCron(env *models.Env, form *forms.DeployEnvForm) e.Error {
	cronDriftParam, err := GetCronDriftParam(forms.CronDriftForm{
		BaseForm:         form.BaseForm,
		CronDriftExpress: form.CronDriftExpress,
		AutoRepairDrift:  form.AutoRepairDrift,
		OpenCronDrift:    form.OpenCronDrift,
	})
	if err != nil {
		return err
	}

	if cronDriftParam.AutoRepairDrift != nil {
		env.AutoRepairDrift = *cronDriftParam.AutoRepairDrift
	}
	if cronDriftParam.OpenCronDrift != nil {
		env.OpenCronDrift = *cronDriftParam.OpenCronDrift
		env.NextDriftTaskTime = cronDriftParam.NextDriftTaskTime
	}
	if cronDriftParam.CronDriftExpress != nil {
		env.CronDriftExpress = *cronDriftParam.CronDriftExpress
	}

	return nil
}

func setAndCheckEnvByForm(c *ctx.ServiceContext, tx *db.Session, env *models.Env, form *forms.DeployEnvForm) e.Error {

	if err := setAndCheckEnvAutoApproval(c, env, form); err != nil {
		return err
	}

	if err := setAndCheckEnvDestroy(env, form); err != nil {
		return err
	}

	if err := setAndCheckEnvCron(env, form); err != nil {
		return err
	}

	if form.HasKey("variables") {
		updateVarsForm := forms.UpdateObjectVarsForm{
			Scope:     consts.ScopeEnv,
			ObjectId:  env.Id,
			Variables: form.Variables,
		}
		if _, er := updateObjectVars(c, tx, &updateVarsForm); er != nil {
			return e.AutoNew(er, e.InternalError)
		}
	}

	if len(form.PolicyGroup) > 0 {
		policyForm := &forms.UpdatePolicyRelForm{
			Id:             env.Id,
			Scope:          consts.ScopeEnv,
			PolicyGroupIds: form.PolicyGroup,
		}
		if _, err := services.UpdatePolicyRel(tx, policyForm); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if form.TaskType == "" {
		return e.New(e.BadParam, http.StatusBadRequest)
	}

	if form.HasKey("varGroupIds") || form.HasKey("delVarGroupIds") {
		// 创建变量组与实例的关系
		if err := services.BatchUpdateRelationship(tx, form.VarGroupIds, form.DelVarGroupIds, consts.ScopeEnv, env.Id.String()); err != nil {
			return err
		}
	}

	return nil
}

func envDeploy(c *ctx.ServiceContext, tx *db.Session, form *forms.DeployEnvForm) (*models.EnvDetail, e.Error) {
	c.AddLogField("action", fmt.Sprintf("deploy env task %s", form.Id))
	if err := envPreCheck(c.OrgId, c.ProjectId, form.KeyId, form.Playbook); err != nil {
		return nil, err
	}

	// 检查自动纠漂移、推送到分支时重新部署时，是否了配置自动审批
	if !services.CheckoutAutoApproval(form.AutoApproval, form.AutoRepairDrift, form.Triggers) {
		return nil, e.New(e.EnvCheckAutoApproval, http.StatusBadRequest)
	}

	// env 检查
	env, err := envCheck(tx, c.OrgId, c.ProjectId, form.Id, c.Logger())
	if err != nil {
		return nil, err
	}

	// 模板检查
	tpl, err := envTplCheck(tx, c.OrgId, env.TplId, c.Logger())
	if err != nil {
		return nil, err
	}

	// set env from form
	setEnvByForm(env, form)

	// set and check autoApproval, destroyAt, cronDrift, TaskType ...
	err = setAndCheckEnvByForm(c, tx, env, form)
	if err != nil {
		return nil, err
	}

	targets := make([]string, 0)
	if len(strings.TrimSpace(form.Targets)) > 0 {
		targets = strings.Split(strings.TrimSpace(form.Targets), ",")
	}

	// 计算变量列表
	vars, er := services.GetValidVarsAndVgVars(tx, env.OrgId, env.ProjectId, env.TplId, env.Id)
	if er != nil {
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
		c.Logger().Errorf("error creating task, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	// Save() 调用会全量将结构体中的字段进行保存，即使字段为 zero value
	if _, err := tx.Save(env); err != nil {
		c.Logger().Errorf("error save env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	env.MergeTaskStatus()
	envDetail := &models.EnvDetail{
		Env:    *env,
		TaskId: task.Id,
	}
	envDetail = PopulateLastTask(c.DB(), envDetail)
	vcs, _ := services.QueryVcsByVcsId(tpl.VcsId, c.DB())
	// 获取token
	token, err := GetWebhookToken(c)
	if err != nil {
		return nil, err
	}

	if err := vcsrv.SetWebhook(vcs, tpl.RepoId, token.Key, form.Triggers); err != nil {
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
		NoPageSizeForm: form.NoPageSizeForm,
		Id:             env.LastResTaskId,
		Q:              form.Q,
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

// EnvVariables 获取环境的变量列表，环境部署对应的环境变量为 last task 固化的变量内容
func EnvVariables(c *ctx.ServiceContext, form forms.SearchEnvVariableForm) (interface{}, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" || form.Id == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	task := models.Task{}
	err := services.QueryWithOrgProject(c.DB(), c.OrgId, c.ProjectId, models.Task{}.TableName()).
		Where("env_id = ? AND `type` IN (?)", form.Id,
			[]string{common.TaskTypePlan, common.TaskTypeApply, common.TaskTypeDestroy}).Last(&task)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.ObjectNotExists, http.StatusNotFound)
		}
		c.Logger().Errorf("query env last task error: %v", err)
		return nil, e.AutoNew(err, e.DBError)
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
func ResourceGraphDetail(c *ctx.ServiceContext, form *forms.ResourceGraphDetailForm) (interface{}, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" || form.Id == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}

	resource, err := services.GetResourceById(c.DB(), form.ResourceId)
	if err != nil {
		c.Logger().Errorf("error get resource, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	res, err := services.GetResourceDetail(c.DB(), c.OrgId, c.ProjectId, form.Id, resource.Id)
	if err != nil {
		c.Logger().Errorf("error get resource, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	if res.EnvId != form.Id || res.OrgId != c.OrgId || res.ProjectId != c.ProjectId {
		c.Logger().Errorf("Environment ID and resource ID do not match")
		return nil, e.New(e.DBError, err, http.StatusForbidden)
	}
	resultAttrs := resource.Attrs
	if len(res.SensitiveKeys) > 0 {
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
	if res.DriftDetail != "" {
		res.IsDrift = true
	}
	res.Attrs = resultAttrs
	return res, nil
}
