// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

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
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"cloudiac/portal/services/vcsrv"
	"cloudiac/utils/logs"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/robfig/cron/v3"

	"github.com/lib/pq"
)

// SpecParser 最小时间单位为分钟
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
		if !autoRepairDrift {
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

func GetNextCronTime(cronExpress string) (*time.Time, e.Error) {
	if cronExpress == "" {
		return nil, e.New(e.BadParam, "cron express is empty")
	}

	return ParseCronpress(cronExpress)
}

func createEnvCheck(c *ctx.ServiceContext, form *forms.CreateEnvForm) e.Error {
	if c.OrgId == "" || c.ProjectId == "" {
		return e.New(e.BadRequest, http.StatusBadRequest)
	}

	// 检查自动纠漂移、推送到分支时重新部署时，是否了配置自动审批
	if !services.CheckoutAutoApproval(form.AutoApproval, form.AutoRepairDrift, form.Triggers) {
		return e.New(e.EnvCheckAutoApproval, http.StatusBadRequest)
	}

	if form.Playbook != "" && form.KeyId == "" {
		return e.New(e.TemplateKeyIdNotSet)
	}

	if er := services.CheckEnvTags(form.Tags); er != nil {
		return er
	}

	return nil
}

//nolint
func setDefaultValueFromTpl(form *forms.CreateEnvForm, tpl *models.Template, destroyAt, deployAt *models.Time, session *db.Session) e.Error {
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

	if !form.HasKey("workdir") {
		form.Workdir = tpl.Workdir
	}

	if !form.HasKey("revision") {
		form.Revision = tpl.RepoRevision
	}

	if !form.HasKey("policyEnable") {
		form.PolicyEnable = tpl.PolicyEnable
	}

	if form.PolicyEnable {
		if !form.HasKey("policyGroup") || len(form.PolicyGroup) == 0 {
			temp, err := services.GetPolicyRels(session, tpl.Id, consts.ScopeTemplate)
			if err != nil {
				return err
			}
			policyGroups := make([]models.Id, 0)
			for _, v := range temp {
				policyGroups = append(policyGroups, models.Id(v.PolicyGroupId))
			}
			form.PolicyGroup = policyGroups
		}
	}

	if !form.HasKey("stepTimeout") || form.StepTimeout == 0 {
		stepTimeout, err := services.GetSystemTaskStepTimeout(session)
		if err != nil {
			return err
		}
		// 以分钟为单位返回
		form.StepTimeout = stepTimeout / 60
	}

	if form.DestroyAt != "" {
		var err error
		*destroyAt, err = models.Time{}.Parse(form.DestroyAt)
		if err != nil {
			return e.New(e.BadParam, http.StatusBadRequest, err)
		}
	} else if form.TTL != "" {
		_, err := services.ParseTTL(form.TTL) // 检查 ttl 格式
		if err != nil {
			return e.New(e.BadParam, http.StatusBadRequest, err)
		}
	} else if form.AutoDestroyCron != "" {
		// 活跃环境同步修改 destroyAt
		at, err := GetNextCronTime(form.AutoDestroyCron)
		if err != nil {
			return err
		}

		mt := models.Time(*at)
		destroyAt = &mt
	}

	if form.DeployAt != "" {
		var err error
		*deployAt, err = models.Time{}.Parse(form.DeployAt)
		if err != nil {
			return e.New(e.BadParam, http.StatusBadRequest, err)
		}
	} else if form.AutoDeployCron != "" {
		// 活跃环境同步修改 deployAt
		at, err := GetNextCronTime(form.AutoDeployCron)
		if err != nil {
			return err
		}

		mt := models.Time(*at)
		deployAt = &mt
	}

	return nil
}

// getTaskStepTimeoutInSecond return timeout in second
func getTaskStepTimeoutInSecond(timeoutInMinute int) (int, e.Error) {
	timeoutInSecond := timeoutInMinute * 60
	if timeoutInSecond <= 0 {
		sysTimeout, err := services.GetSystemTaskStepTimeout(db.Get())
		if err != nil {
			return -1, err
		}
		timeoutInSecond = sysTimeout
	}
	return timeoutInSecond, nil
}

func createEnvToDB(tx *db.Session, c *ctx.ServiceContext, form *forms.CreateEnvForm, envModel models.Env) (*models.Env, e.Error) {
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

	return env, nil
}

func handlerCreateEnvVars(tx *db.Session, c *ctx.ServiceContext, form *forms.CreateEnvForm, env *models.Env) ([]string, []models.VariableBody, e.Error) {
	// sampleVariables是外部下发的环境变量，会和计算出来的变量列表冲突，这里需要做一下处理
	// FIXME 未对变量组的变量进行处理
	sampleVars, err := services.GetSampleValidVariables(tx, c.OrgId, c.ProjectId, env.TplId, env.Id, form.SampleVariables)
	if err != nil {
		return nil, nil, e.New(err.Code(), err, http.StatusInternalServerError)
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
		return nil, nil, e.AutoNew(er, e.InternalError)
	}

	// 创建变量组与实例的关系
	if err := services.BatchUpdateVarGroupObjectRel(tx, form.VarGroupIds, form.DelVarGroupIds, consts.ScopeEnv, env.Id); err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	targets := make([]string, 0)
	if len(strings.TrimSpace(form.Targets)) > 0 {
		targets = strings.Split(strings.TrimSpace(form.Targets), ",")
	}

	// 计算变量列表
	vars, er := services.GetValidVarsAndVgVars(tx, env.OrgId, env.ProjectId, env.TplId, env.Id)
	if er != nil {
		_ = tx.Rollback()
		return nil, nil, err
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
			return nil, nil, err
		}
	}

	return targets, vars, nil
}

func getCreateEnvTpl(c *ctx.ServiceContext, form *forms.CreateEnvForm) (*models.Template, e.Error) {
	query := c.DB().Where("status = ?", models.Enable)
	tpl, err := services.GetTemplateById(query, form.TplId)
	if err != nil && err.Code() == e.TemplateNotExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error get template, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}
	return tpl, nil
}

func envWorkdirCheck(c *ctx.ServiceContext, repoId, repoRevision, workdir string, vcsId models.Id) e.Error {
	searchForm := &forms.RepoFileSearchForm{
		RepoId:       repoId,
		RepoRevision: repoRevision,
		VcsId:        vcsId,
		Workdir:      workdir,
	}
	results, err := VcsRepoFileSearch(c, searchForm, "", consts.TfFileMatch)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		return e.New(e.TemplateWorkdirError, fmt.Errorf("no '%s' files", consts.TfFileMatch))
	}
	return nil
}

// CreateEnv 创建环境
// nolint:cyclop
func CreateEnv(c *ctx.ServiceContext, form *forms.CreateEnvForm) (*models.EnvDetail, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create env %s", form.Name))

	if form.KeyId == "" && form.KeyName != "" {
		query := services.QueryKey(services.QueryWithOrgId(c.DB(), c.OrgId))
		if key, _ := services.GetKeyByName(query, form.KeyName); key != nil {
			form.KeyId = key.Id
		}
	}
	err := createEnvCheck(c, form)
	if err != nil {
		return nil, err
	}

	err = services.IsTplAssociationCurrentProject(c, form.TplId)
	if err != nil {
		return nil, err
	}

	// 检查模板
	tpl, err := getCreateEnvTpl(c, form)
	if err != nil {
		return nil, err
	}

	// 以下值只在未传入时使用模板定义的值，如果入参有该字段即使值为空也不会使用模板中的值
	var (
		destroyAt models.Time
		deployAt  models.Time
	)
	err = setDefaultValueFromTpl(form, tpl, &destroyAt, &deployAt, c.DB())
	if err != nil {
		return nil, err
	}

	// 检查环境传入工作目录
	if err = envWorkdirCheck(c, tpl.RepoId, form.Revision, form.Workdir, tpl.VcsId); err != nil {
		return nil, err
	}

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 同组织同项目环境不允许重名
	if env, _ := services.GetEnvByName(tx, c.OrgId, c.ProjectId, form.Name); env != nil {
		return nil, e.New(e.EnvAlreadyExists, http.StatusBadRequest)
	}

	taskStepTimeout, err := getTaskStepTimeoutInSecond(form.StepTimeout)
	if err != nil {
		return nil, err
	}

	envModel := models.Env{
		OrgId:     c.OrgId,
		ProjectId: c.ProjectId,
		CreatorId: c.UserId,
		TokenId:   c.ApiTokenId,
		TplId:     form.TplId,

		Name:        form.Name,
		Tags:        strings.TrimSpace(form.Tags),
		RunnerId:    form.RunnerId,
		RunnerTags:  strings.Join(form.RunnerTags, ","),
		Status:      models.EnvStatusInactive,
		OneTime:     form.OneTime,
		StepTimeout: taskStepTimeout,

		// 模板参数
		TfVarsFile:   form.TfVarsFile,
		PlayVarsFile: form.PlayVarsFile,
		Playbook:     form.Playbook,
		Revision:     form.Revision,
		KeyId:        form.KeyId,
		Workdir:      form.Workdir,

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

		AutoDeployAt:    &deployAt,
		AutoDeployCron:  form.AutoDeployCron,
		AutoDestroyCron: form.AutoDestroyCron,
	}

	if tpl.IsDemo {
		// 演示环境强制设置自动销毁
		envModel.IsDemo = true
		envModel.TTL = consts.DemoEnvTTL
		envModel.AutoDestroyAt = nil
		envModel.AutoApproval = true
	}

	env, err := createEnvToDB(tx, c, form, envModel)
	if err != nil {
		return nil, err
	}

	targets, vars, err := handlerCreateEnvVars(tx, c, form, env)
	if err != nil {
		return nil, err
	}

	// 来源：手动触发、外部调用
	taskSource, taskSourceSys := getEnvSource(form.Source)

	if _, er := services.UpdateObjectTags(tx, c.OrgId, env.Id,
		consts.ScopeEnv, consts.TagSourceApi, tagList2Map(form.EnvTags)); er != nil {
		return nil, er
	}
	if _, er := services.UpdateObjectTags(tx, c.OrgId, env.Id,
		consts.ScopeEnv, consts.TagSourceUser, tagList2Map(form.UserTags)); er != nil {
		return nil, err
	}

	// 创建任务
	task, err := services.CreateTask(tx, tpl, env, models.Task{
		Name:            models.Task{}.GetTaskNameByType(form.TaskType),
		Targets:         targets,
		CreatorId:       c.UserId,
		TokenId:         c.ApiTokenId,
		KeyId:           env.KeyId,
		Variables:       vars,
		AutoApprove:     env.AutoApproval,
		Revision:        env.Revision,
		StopOnViolation: env.StopOnViolation,
		BaseTask: models.BaseTask{
			Type:        form.TaskType,
			StepTimeout: taskStepTimeout,
			RunnerId:    env.RunnerId,
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
	if _, err := tx.UpdateAll(env); err != nil {
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
		EnvTags:    form.EnvTags,
		UserTags:   form.UserTags,
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

	// 记录操作日志
	services.InsertUserOperateLog(c.UserId, c.OrgId, env.Id, consts.OperatorObjectTypeEnv, "create", env.Name, nil)

	return &envDetail, nil
}

// SearchEnv 环境查询
func SearchEnv(c *ctx.ServiceContext, form *forms.SearchEnvForm) (interface{}, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	//query := c.DB().Where("iac_env.org_id = ? AND iac_env.project_id = ?", c.OrgId, c.ProjectId)
	query := services.QueryEnvDetail(c.DB(), c.OrgId, c.ProjectId)
	var er e.Error

	// 环境状态过滤
	query, er = services.FilterEnvStatus(query, form.Status, form.Deploying)
	if er != nil {
		return nil, er
	}

	// 环境归档状态过滤
	query, er = services.FilterEnvArchiveStatus(query, form.Archived)
	if er != nil {
		return nil, er
	}

	// 环境更新时间过滤
	query = services.FilterEnvUpdatedTime(query, form.StartTime, form.EndTime)

	if form.Q != "" {
		qs := "%" + form.Q + "%"
		query = query.Joins("left join iac_template on iac_env.tpl_id = iac_template.id")
		query = query.Where("iac_env.name LIKE ? OR iac_template.name LIKE ? OR iac_env.tags LIKE ?", qs, qs, qs)
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

	enabledBill, err := services.ProjectEnabledBill(c.DB(), c.ProjectId)
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	for _, env := range details {
		env.MergeTaskStatus()
		PopulateLastTask(c.DB(), env)
		env.PolicyStatus = models.PolicyStatusConversion(env.PolicyStatus, env.PolicyEnable)
		// runner tags 数组形式返回
		if env.Env.RunnerTags != "" {
			env.RunnerTags = strings.Split(env.Env.RunnerTags, ",")
		} else {
			env.RunnerTags = []string{}
		}
		// 以分钟为单位返回
		env.StepTimeout = env.StepTimeout / 60

		// 是否开启费用采集
		env.IsBilling = enabledBill

		// 标签
		tags, _ := services.FindObjectTags(db.Get(), c.OrgId, env.Id, consts.ScopeEnv)
		for _, t := range tags {
			if t.Source == consts.TagSourceApi {
				env.EnvTags = append(env.EnvTags, models.Tag{Key: t.Key, Value: t.Value})
			} else {
				env.UserTags = append(env.UserTags, models.Tag{Key: t.Key, Value: t.Value})
			}
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

			if token, _ := services.GetApiTokenByIdRaw(query, lastTask.TokenId); token != nil {
				env.TokenName = token.Name
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
	if form.HasKey("stepTimeout") {
		// 将分钟转换为秒
		attrs["stepTimeout"] = form.StepTimeout * 60
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

func setAndCheckUpdateEnvDeploy(tx *db.Session, attrs models.Attrs, env *models.Env, form *forms.UpdateEnvForm) e.Error {
	if !form.HasKey("deployAt") && !form.HasKey("autoDeployCron") {
		return nil
	}

	if form.HasKey("deployAt") {
		deployAt, err := models.Time{}.Parse(form.DeployAt)
		if err != nil {
			_ = tx.Rollback()
			return e.New(e.BadParam, http.StatusBadRequest, err)
		}
		attrs["auto_deploy_at"] = &deployAt
		attrs["auto_deploy_cron"] = ""
	}
	if form.HasKey("autoDeployCron") {
		attrs["auto_deploy_at"] = nil
		attrs["auto_deploy_cron"] = form.AutoDeployCron

		if form.AutoDeployCron != "" {
			// 活跃环境同步修改 deployAt
			at, err := GetNextCronTime(form.AutoDeployCron)
			if err != nil {
				return err
			}

			mt := models.Time(*at)
			attrs["auto_deploy_at"] = &mt
		}
	}
	return nil
}

func setAndCheckUpdateEnvDestroy(tx *db.Session, attrs models.Attrs, env *models.Env, form *forms.UpdateEnvForm) e.Error {
	if !form.HasKey("destroyAt") && !form.HasKey("ttl") && !form.HasKey("autoDestroyCron") {
		return nil
	}

	if form.HasKey("destroyAt") {
		destroyAt, err := models.Time{}.Parse(form.DestroyAt)
		if err != nil {
			_ = tx.Rollback()
			return e.New(e.BadParam, http.StatusBadRequest, err)
		}
		attrs["auto_destroy_at"] = &destroyAt
		attrs["ttl"] = "" // 直接传入了销毁时间，需要同步清空 ttl
		attrs["auto_destroy_cron"] = ""
		attrs["auto_deploy_cron"] = ""
		attrs["auto_deploy_at"] = nil
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
		} else if env.Status != models.EnvStatusDestroyed && env.Status != models.EnvStatusInactive {
			// 活跃环境同步修改 destroyAt
			at := models.Time(time.Now().Add(ttl))
			attrs["auto_destroy_at"] = &at
		}
		attrs["auto_destroy_cron"] = ""
		attrs["auto_deploy_cron"] = ""
		attrs["auto_deploy_at"] = nil
	}

	if form.HasKey("autoDestroyCron") {
		attrs["ttl"] = ""
		attrs["auto_destroy_at"] = nil
		attrs["auto_destroy_cron"] = form.AutoDestroyCron

		if form.AutoDestroyCron != "" {
			// 活跃环境同步修改 destroyAt
			at, err := GetNextCronTime(form.AutoDestroyCron)
			if err != nil {
				return err
			}

			mt := models.Time(*at)
			attrs["auto_destroy_at"] = &mt
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
			c.Logger().Errorf("set webhook err :%v", err)
		}
	}
	return nil
}

func setAndCheckUpdateEnvByForm(c *ctx.ServiceContext, tx *db.Session, attrs models.Attrs, env *models.Env, form *forms.UpdateEnvForm) e.Error { // nolint:cyclop
	if form.HasKey("tags") {
		if er := services.CheckEnvTags(form.Tags); er != nil {
			return er
		} else {
			attrs["tags"] = strings.TrimSpace(form.Tags)
		}
	}

	if !env.IsDemo { // 演示环境不允许修改自动审批和存活时间
		if err := setAndCheckUpdateEnvAutoApproval(c, tx, attrs, env, form); err != nil {
			return err
		}

		if err := setAndCheckUpdateEnvDestroy(tx, attrs, env, form); err != nil {
			return err
		}

		if err := setAndCheckUpdateEnvDeploy(tx, attrs, env, form); err != nil {
			return err
		}
	}

	if err := setAndCheckUpdateEnvTriggers(c, tx, attrs, env, form); err != nil {
		return err
	}

	if form.HasKey("archived") {
		envResCount := int64(0)
		if env.LastResTaskId != "" {
			var err e.Error
			envResCount, err = services.GetTaskResourceCount(tx, env.LastResTaskId)
			if err != nil {
				return err
			}
		}
		if !(env.Status == models.EnvStatusInactive ||
			env.Status == models.EnvStatusDestroyed ||
			(env.Status == models.EnvStatusFailed && envResCount == 0)) {
			_ = tx.Rollback()
			return e.New(e.EnvCannotArchiveActive,
				fmt.Errorf("env can't be archive while env is %s", env.Status),
				http.StatusBadRequest)
		}
		attrs["archived"] = form.Archived
		if form.Name != "" {
			attrs["name"] = form.Name
		}
	}
	return nil
}

// UpdateEnv 环境编辑
func UpdateEnv(c *ctx.ServiceContext, form *forms.UpdateEnvForm) (*models.EnvDetail, e.Error) { // nolint:cyclop
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
		_ = tx.Rollback()
		return nil, err
	}
	if env.Locked {
		_ = tx.Rollback()
		return nil, e.New(e.EnvLocked, http.StatusBadRequest)
	}
	if !env.Archived {
		if form.Archived {
			// 环境归档时自动重新命名
			form.Name = env.Name + "-archived-" + time.Now().Format("20060102150405")
		}
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

	if form.HasKey("keyName") {
		query := services.QueryKey(services.QueryWithOrgId(c.DB(), c.OrgId))
		if key, _ := services.GetKeyByName(query, form.KeyName); key != nil {
			form.KeyId = key.Id
		}
	}

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

	// 记录操作日志
	services.InsertUserOperateLog(c.UserId, c.OrgId, env.Id, consts.OperatorObjectTypeEnv, "update", form.Name, nil)

	return detail, nil
}

// EnvDetail 环境信息详情
func EnvDetail(c *ctx.ServiceContext, form forms.DetailEnvForm) (*models.EnvDetail, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	//query := c.DB().Where("iac_env.org_id = ? AND iac_env.project_id = ?", c.OrgId, c.ProjectId)
	query := services.QueryEnvDetail(c.DB(), c.OrgId, c.ProjectId)

	envDetail, err := services.GetEnvDetailById(query, form.Id)
	if err != nil && err.Code() == e.EnvNotExists {
		return nil, e.New(e.EnvNotExists, err, http.StatusNotFound)
	} else if err != nil {
		c.Logger().Errorf("error get env by id, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	enabledBill, err := services.ProjectEnabledBill(c.DB(), envDetail.ProjectId)
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
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
	// 时间转化为分钟
	envDetail.StepTimeout = envDetail.StepTimeout / 60

	// runner tags 数组形式返回
	if envDetail.Env.RunnerTags != "" {
		envDetail.RunnerTags = strings.Split(envDetail.Env.RunnerTags, ",")
	} else {
		envDetail.RunnerTags = []string{}
	}
	// 是否开启费用采集
	envDetail.IsBilling = enabledBill

	return envDetail, nil
}

// EnvDeploy 创建新部署任务
// 任务类型：plan, apply, destroy
func EnvDeploy(c *ctx.ServiceContext, form *forms.DeployEnvForm) (ret *models.EnvDetail, er e.Error) {
	_ = c.DB().Transaction(func(tx *db.Session) error {
		ret, er = envDeploy(c, tx, form)
		return er
	})

	// 记录操作日志
	services.InsertUserOperateLog(c.UserId, c.OrgId, form.Id, consts.OperatorObjectTypeEnv, form.TaskType, form.Name, nil)

	return ret, er
}

// EnvDeployCheck 创建新部署前检测
func EnvDeployCheck(c *ctx.ServiceContext, envId models.Id) (interface{}, e.Error) {
	if c.OrgId == "" || c.ProjectId == "" {
		return nil, e.New(e.BadRequest, http.StatusBadRequest)
	}
	env, err := services.GetEnvById(c.DB(), envId)
	if err != nil {
		return nil, err
	}
	//判断环境是否已归档
	if env.Archived {
		return nil, e.New(e.EnvArchived, "Environment archived")
	}

	// 云模板检测
	tpl, err := services.GetTplByEnvId(c.DB(), envId)
	if err != nil {
		return nil, err
	}
	//检测云模板是否绑定项目
	_, checkErr := services.GetBindTemplate(c.DB(), c.ProjectId, tpl.Id)
	if checkErr != nil {
		return nil, checkErr
	}
	//vcs 检测(是否禁用，token是否有效)
	vcs, err := services.GetVcsById(c.DB(), tpl.VcsId)
	if err != nil {
		return nil, err
	}
	if vcs.Status != "enable" {
		return nil, e.New(e.VcsError, "vcs is disable")
	}
	if err := services.VscTokenCheckByID(c.DB(), vcs.Id, vcs.VcsToken); err != nil {
		return nil, e.New(e.VcsInvalidToken, err)
	}
	//环境运行中不允许再手动发布任务
	tasks, err := services.GetActiveTaskByEnvId(c.DB(), envId)
	if err != nil {
		return nil, err
	}
	if len(tasks) > 0 {
		return nil, e.New(e.EnvDeploying, "Deployment initiation is not allowed")
	}
	return nil, nil
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

	if form.HasKey("stepTimeout") {
		// 将分钟转换为秒
		env.StepTimeout = form.StepTimeout * 60
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
	if form.HasKey("workdir") {
		env.Workdir = form.Workdir
	}

	setEnvRunnerInfoByForm(env, form)
}

func setEnvRunnerInfoByForm(env *models.Env, form *forms.DeployEnvForm) {
	if form.HasKey("runnerId") {
		env.RunnerId = form.RunnerId
	}

	if form.HasKey("runnerTags") {
		env.RunnerTags = strings.Join(form.RunnerTags, ",")
		// 如果传了 tags 则清空 runnerId 值
		env.RunnerId = ""
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

func setAndCheckEnvAutoDeploy(tx *db.Session, env *models.Env, form *forms.DeployEnvForm) e.Error { //nolint:dupl
	if !form.HasKey("deployAt") && !form.HasKey("autoDeployCron") {
		return nil
	}

	if form.HasKey("deployAt") {
		deployAt, err := models.Time{}.Parse(form.DeployAt)
		if err != nil {
			return e.New(e.BadParam, http.StatusBadRequest, err)
		}
		env.AutoDeployAt = &deployAt
		// 直接传入了部署时间，需要同步清空 ttl
		env.AutoDeployCron = ""
	}
	if form.HasKey("autoDeployCron") {
		env.AutoDeployAt = nil
		env.AutoDeployCron = form.AutoDeployCron

		if env.AutoDeployCron != "" {
			// 非活跃环境同步修改 deployAt
			at, err := GetNextCronTime(env.AutoDeployCron)
			if err != nil {
				return err
			}

			mt := models.Time(*at)
			env.AutoDeployAt = &mt
		}
	}

	return nil
}

func setAndCheckEnvAutoDestroy(tx *db.Session, env *models.Env, form *forms.DeployEnvForm) e.Error { //nolint:dupl
	if !form.HasKey("destroyAt") && !form.HasKey("ttl") && !form.HasKey("autoDestroyCron") {
		return nil
	}

	if form.HasKey("destroyAt") {
		destroyAt, err := models.Time{}.Parse(form.DestroyAt)
		if err != nil {
			return e.New(e.BadParam, http.StatusBadRequest, err)
		}
		env.AutoDestroyAt = &destroyAt
		// 直接传入了销毁时间，需要同步清空 ttl
		env.TTL = ""
		env.AutoDestroyCron = ""
	} else if form.HasKey("ttl") {
		ttl, err := services.ParseTTL(form.TTL)
		if err != nil {
			return e.New(e.BadParam, http.StatusBadRequest, err)
		}

		env.TTL = form.TTL

		if ttl == 0 { // ttl 传入 0 表示清空自动销毁时间
			env.AutoDestroyAt = nil
		} else if env.Status != models.EnvStatusDestroyed && env.Status != models.EnvStatusInactive {
			// 活跃环境同步修改 destroyAt
			at := models.Time(time.Now().Add(ttl))
			env.AutoDestroyAt = &at
		}
		env.AutoDestroyCron = ""
	}
	if form.HasKey("autoDestroyCron") {
		env.TTL = ""
		env.AutoDestroyAt = nil
		env.AutoDestroyCron = form.AutoDestroyCron

		if env.AutoDestroyCron != "" {
			// 活跃环境同步修改 destroyAt
			at, err := GetNextCronTime(env.AutoDestroyCron)
			if err != nil {
				return err
			}

			mt := models.Time(*at)
			env.AutoDestroyAt = &mt
		}
	}

	return nil
}

func setAndCheckEnvDriftCron(env *models.Env, form *forms.DeployEnvForm) e.Error {
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
	if !env.IsDemo { // 演示环境不允许修改自动审批和自动销毁设置
		if err := setAndCheckEnvAutoApproval(c, env, form); err != nil {
			return err
		}
		if err := setAndCheckEnvAutoDestroy(tx, env, form); err != nil {
			return err
		}
		if err := setAndCheckEnvAutoDeploy(tx, env, form); err != nil {
			return err
		}
	}

	// drift Cron 相关
	if err := setAndCheckEnvDriftCron(env, form); err != nil {
		return err
	}
	if form.HasKey("extraData") {
		env.ExtraData = form.ExtraData
	}

	if form.HasKey("variables") {
		updateVarsForm := forms.UpdateObjectVarsForm{
			Scope:     consts.ScopeEnv,
			ObjectId:  env.Id,
			Variables: checkDeployVar(form.Variables),
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
		if err := services.BatchUpdateVarGroupObjectRel(tx, form.VarGroupIds, form.DelVarGroupIds, consts.ScopeEnv, env.Id); err != nil {
			return err
		}
	}

	return nil
}

func envDeploy(c *ctx.ServiceContext, tx *db.Session, form *forms.DeployEnvForm) (*models.EnvDetail, e.Error) { // nolint:cyclop
	c.AddLogField("action", fmt.Sprintf("deploy env task %s", form.Id))
	lg := c.Logger()

	if form.HasKey("keyName") {
		query := services.QueryKey(services.QueryWithOrgId(c.DB(), c.OrgId))
		if key, _ := services.GetKeyByName(query, form.KeyName); key != nil {
			form.KeyId = key.Id
		}
	}
	if err := envPreCheck(c.OrgId, c.ProjectId, form.KeyId, form.Playbook); err != nil {
		return nil, err
	}
	lg.Debugln("envDeploy -> envPreCheck finish")

	// 检查自动纠漂移、推送到分支时重新部署时，是否了配置自动审批
	if !services.CheckoutAutoApproval(form.AutoApproval, form.AutoRepairDrift, form.Triggers) {
		return nil, e.New(e.EnvCheckAutoApproval, http.StatusBadRequest)
	}
	lg.Debugln("envDeploy -> CheckoutAutoApproval finish")

	// env 检查
	env, err := envCheck(tx, c.OrgId, c.ProjectId, form.Id, c.Logger())
	if err != nil {
		return nil, err
	}
	lg.Debugln("envDeploy -> envCheck finish")

	if form.TaskType != common.TaskTypePlan && env.Locked {
		return nil, e.New(e.EnvLocked, http.StatusBadRequest)
	}

	// 模板检查
	tpl, err := envTplCheck(tx, c.OrgId, env.TplId, c.Logger())
	if err != nil {
		return nil, err
	}

	if !form.HasKey("workdir") {
		form.Workdir = env.Workdir
	}

	if !form.HasKey("revision") {
		form.Revision = env.Revision
	}

	if !form.HasKey("keyId") {
		form.KeyId = env.KeyId
	}

	// 环境下云模版工作目录检查
	if err = envWorkdirCheck(c, tpl.RepoId, form.Revision, form.Workdir, tpl.VcsId); err != nil {
		return nil, err
	}
	lg.Debugln("envDeploy -> envTplCheck finish")

	// set env from form
	setEnvByForm(env, form)

	// set and check autoApproval, destroyAt, cronDrift, TaskType, variables...
	err = setAndCheckEnvByForm(c, tx, env, form)
	if err != nil {
		return nil, err
	}
	lg.Debugln("envDeploy -> setAndCheckEnvByForm finish")

	if env.IsDemo && env.Status == models.EnvStatusDestroyed {
		// 演示环境销毁后重新部署也强制设置自动销毁
		env.TTL = consts.DemoEnvTTL
		env.AutoDestroyAt = nil
		env.AutoApproval = true
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
	lg.Debugln("envDeploy -> GetValidVarsAndVgVars finish")

	// 获取实际执行任务的runnerID
	rId, err := services.GetAvailableRunnerIdByStr(env.RunnerId, env.RunnerTags)
	if err != nil {
		return nil, err
	}

	// 来源：手动触发、外部调用
	taskSource, taskSourceSys := getEnvSource(form.Source)

	// 是否是漂移检测任务
	var IsDriftTask bool
	if form.IsDriftTask && form.TaskType == common.TaskTypePlan {
		IsDriftTask = true
	} else {
		IsDriftTask = false
	}

	if _, er := services.UpdateObjectTags(tx, c.OrgId, env.Id,
		consts.ScopeEnv, consts.TagSourceApi, tagList2Map(form.EnvTags)); er != nil {
		return nil, er
	}
	if _, er := services.UpdateObjectTags(tx, c.OrgId, env.Id,
		consts.ScopeEnv, consts.TagSourceUser, tagList2Map(form.UserTags)); er != nil {
		return nil, err
	}

	// 创建任务
	task, err := services.CreateTask(tx, tpl, env, models.Task{
		Name:            models.Task{}.GetTaskNameByType(form.TaskType),
		Targets:         targets,
		CreatorId:       c.UserId,
		TokenId:         c.ApiTokenId,
		KeyId:           env.KeyId,
		Variables:       vars,
		AutoApprove:     env.AutoApproval,
		Revision:        env.Revision,
		StopOnViolation: env.StopOnViolation,
		ExtraData:       env.ExtraData,
		BaseTask: models.BaseTask{
			Type:        form.TaskType,
			StepTimeout: env.StepTimeout,
			RunnerId:    rId,
		},
		Source:      taskSource,
		SourceSys:   taskSourceSys,
		Callback:    env.Callback,
		IsDriftTask: IsDriftTask,
	})

	if err != nil {
		c.Logger().Errorf("error creating task, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}
	lg.Debugln("envDeploy -> CreateTask finish")

	if _, err := tx.UpdateAll(env); err != nil {
		c.Logger().Errorf("error save env, err %s", err)
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	env.MergeTaskStatus()
	envDetail := &models.EnvDetail{
		Env:      *env,
		TaskId:   task.Id,
		EnvTags:  form.EnvTags,
		UserTags: form.UserTags,
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
		for k := range resultAttrs {
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
		for k := range resultAttrs {
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

func EnvUpdateTags(c *ctx.ServiceContext, form *forms.UpdateEnvTagsForm) (resp interface{}, er e.Error) {
	if er := services.CheckEnvTags(form.Tags); er != nil {
		return nil, er
	}

	query := services.QueryWithOrgProject(c.DB(), c.OrgId, c.ProjectId)
	tags := strings.TrimSpace(form.Tags)
	if env, er := services.UpdateEnv(query, form.Id, models.Attrs{"tags": tags}); er != nil {
		return nil, er
	} else {
		return env, nil
	}
}

func EnvLock(c *ctx.ServiceContext, form *forms.EnvLockForm) (interface{}, e.Error) {
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// 查询环境下是否有执行中、待审批、排队中的任务
	tasks, err := services.GetActiveTaskByEnvId(tx, form.Id)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if len(tasks) > 0 {
		_ = tx.Rollback()
		return nil, e.New(e.EnvLockFailedTaskActive)
	}

	env, err := services.GetEnvDetailById(tx, form.Id)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if env.IsDemo {
		_ = tx.Rollback()
		return nil, e.New(e.EnvLockedFailedEnvIsDemo)
	}

	if err := services.EnvLock(tx, form.Id); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}
	return nil, nil
}

func EnvUnLock(c *ctx.ServiceContext, form *forms.EnvUnLockForm) (interface{}, e.Error) {
	attrs := models.Attrs{}
	attrs["locked"] = false

	if form.ClearDestroyAt {
		attrs["auto_destroy_at"] = nil
		attrs["ttl"] = ""
	}

	if _, err := services.UpdateEnv(c.DB(), form.Id, attrs); err != nil {
		return nil, err
	}
	return nil, nil
}

func EnvUnLockConfirm(c *ctx.ServiceContext, form *forms.EnvUnLockConfirmForm) (interface{}, e.Error) {
	env, err := services.GetEnvById(c.DB(), form.Id)
	if err != nil {
		return nil, err
	}

	resp := resps.EnvUnLockConfirmResp{}
	if env.AutoDestroyAt != nil && time.Now().Unix() > env.AutoDestroyAt.Unix() && env.AutoDestroyAt.Unix() > 0 {
		resp.AutoDestroyPass = true
	}

	return resp, nil
}

// EnvStat 环境概览页统计数据
func EnvStat(c *ctx.ServiceContext, form *forms.EnvParam) (interface{}, e.Error) {

	tx := c.DB()

	// 费用类型统计
	envCostTypeStat, err := services.EnvCostTypeStat(tx, form.Id)
	if err != nil {
		return nil, err
	}

	// 费用趋势统计
	envCostTrendStat, err := services.EnvCostTrendStat(tx, form.Id, 6)
	if err != nil {
		return nil, err
	}

	// 费用列表
	envCostList, err := services.EnvCostList(tx, form.Id)
	if err != nil {
		return nil, err
	}

	var results = make([]resps.EnvCostDetailResp, 0)
	for _, envCost := range envCostList {
		resInfo := GetCloudResourceInfo(envCost.Attrs, envCost.ResType)

		results = append(results, resps.EnvCostDetailResp{
			ResType:          envCost.ResType,
			ResAttr:          GetResShowName(envCost.Attrs, envCost.Address),
			InstanceId:       envCost.InstanceId,
			CurMonthCost:     envCost.CurMonthCost,
			TotalCost:        envCost.TotalCost,
			InstanceSpec:     resInfo[consts.InstanceSpecKey],
			SubscriptionType: resInfo[consts.SubscriptionTypeKey],
			Region:           resInfo[consts.RegionKey],
			AvailabilityZone: resInfo[consts.ZoneKey],
		})
	}

	return &resps.EnvStatisticsResp{
		CostTypeStat:  envCostTypeStat,
		CostTrendStat: envCostTrendStat,
		CostList:      results,
	}, nil
}

func checkDeployVar(vars []forms.Variable) []forms.Variable {
	resp := make([]forms.Variable, 0)
	for _, v := range vars {
		if v.Scope != consts.ScopeEnv {
			continue
		}
		resp = append(resp, v)
	}

	return resp
}

func getEnvSource(source string) (taskSource string, taskSourceSys string) {
	taskSource = consts.TaskSourceManual
	taskSourceSys = ""
	if source != consts.TaskSourceManual {
		taskSource = consts.TaskSourceApi
		taskSourceSys = source
	}
	return
}

func GetCloudResourceInfo(attrs map[string]interface{}, resType string) map[string]string {
	var subscriptionFuncs = map[string]consts.SubscriptionFunc{
		consts.AliCloudInstance:    getAliyunInstanceSubscriptionType,
		consts.AliCloudSLB:         getAliyunPaymentSubscriptionType,
		consts.AliCloudALB:         getAliyunPaymentSubscriptionType,
		consts.AliCloudSLBClassic:  getAliyunPaymentSubscriptionType,
		consts.AliCloudDisk:        getAliyunPaymentSubscriptionType,
		consts.AliCloudDiskClassic: getAliyunPaymentSubscriptionType,
		consts.AliCloudEIP:         getAliyunPaymentSubscriptionType,
		consts.AliCloudDB:          getAliyunInstanceSubscriptionType,
		consts.AliCloudMongoDB:     getAliyunInstanceSubscriptionType,
		consts.AliCloudKVStore:     getAliyunPaymentSubscriptionType,
	}

	result := make(map[string]string)

	subscriptionTypeFunc, ok := subscriptionFuncs[resType]
	if !ok {
		return result
	}

	result[consts.ZoneKey] = getStringValue(attrs, getZoneKey(resType))
	result[consts.RegionKey] = getRegionFromAvailabilityZone(getStringValue(attrs, getZoneKey(resType)))
	result[consts.InstanceSpecKey] = getStringValue(attrs, getSpecKey(resType))
	result[consts.SubscriptionTypeKey] = subscriptionTypeFunc(attrs)

	return result
}

func getRegionFromAvailabilityZone(availabilityZone string) string {
	// find the index position of the last "-"
	lastDashIndex := strings.LastIndex(availabilityZone, "-")
	// if "-" is not found, return the original string directly
	if lastDashIndex == -1 {
		return availabilityZone
	}

	// intercept the content after the last "-"
	suffix := availabilityZone[lastDashIndex+1:]

	// whether there is a number after the last "-"
	numLen := 0
	for _, c := range suffix {
		if unicode.IsDigit(c) {
			numLen++
		} else {
			break
		}
	}

	// if it contains a number, return the content before the last "-" and the number part
	if numLen > 0 {
		return availabilityZone[:lastDashIndex+1] + suffix[:numLen]
	}

	// if it does not contain a number, directly return the content before the last "-"
	return availabilityZone[:lastDashIndex]
}

func getZoneKey(resType string) string {
	switch resType {
	case consts.AliCloudInstance, consts.AliCloudDisk, consts.AliCloudDiskClassic:
		return consts.ZoneKey
	case consts.AliCloudSLB, consts.AliCloudSLBClassic, consts.AliCloudALB:
		return consts.SLBZoneKey
	case consts.AliCloudDB, consts.AliCloudMongoDB, consts.AliCloudKVStore:
		return consts.ZoneIdKey
	case consts.AliCloudEIP:
		return ""
	default:
		return ""
	}
}

func getSpecKey(resType string) string {
	switch resType {
	case consts.AliCloudInstance, consts.AliCloudDB:
		return consts.InstanceTypeKey
	case consts.AliCloudSLB, consts.AliCloudSLBClassic, consts.AliCloudALB:
		return consts.SpecificationKey
	case consts.AliCloudDisk, consts.AliCloudDiskClassic:
		return consts.CategoryKey
	case consts.AliCloudMongoDB:
		return consts.MongoDBTypeKey
	case consts.AliCloudKVStore:
		return consts.KVStoreTypeKey
	case consts.AliCloudEIP:
		return ""
	default:
		return ""
	}
}

func getAliyunInstanceSubscriptionType(attrs map[string]interface{}) string {
	if getStringValue(attrs, consts.ChargeTypeKey) == consts.PrePaid {
		return "Subscription"
	}

	if getStringValue(attrs, consts.ChargeTypeKey) == consts.PostPaid {
		if v, ok := attrs[consts.SpotStrategyKey]; ok && v.(string) != "" && v.(string) != "NoSpot" {
			return "Spot"
		}
		return "PayAsYouGo"
	}

	return ""
}

func getAliyunPaymentSubscriptionType(attrs map[string]interface{}) string {
	return getStringValue(attrs, consts.PaymentTypeKey)
}

func getStringValue(attrs map[string]interface{}, key string) string {
	if v, ok := attrs[key]; ok {
		return v.(string)
	}
	return ""
}

func tagList2Map(tags []models.Tag) map[string]string {
	rv := make(map[string]string)
	for _, t := range tags {
		rv[t.Key] = t.Value
	}
	return rv
}
