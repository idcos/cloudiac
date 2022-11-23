// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/libs/ctx"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"

	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/utils"
	"cloudiac/utils/logs"
)

func GetEnv(sess *db.Session, id models.Id) (*models.Env, error) {
	env := models.Env{}
	err := sess.Where("id = ?", id).First(&env)
	return &env, err
}

func CreateEnv(tx *db.Session, env models.Env) (*models.Env, e.Error) {
	if env.Id == "" {
		env.Id = models.NewId("env")
	}
	if env.StatePath == "" {
		env.StatePath = env.DefaultStatPath()
	}
	if err := models.Create(tx, &env); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.EnvNameDuplicated, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &env, nil
}

func UpdateEnv(tx *db.Session, id models.Id, attrs models.Attrs) (env *models.Env, re e.Error) {
	env = &models.Env{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Env{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.EnvAliasDuplicate)
		}
		return nil, e.New(e.DBError, fmt.Errorf("update env error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(env); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query env error: %v", err))
	}
	return
}

func UpdateEnvModel(tx *db.Session, id models.Id, env models.Env) e.Error {
	_, err := models.UpdateModel(tx.Where("id = ?", id), &env)
	if err != nil {
		return e.AutoNew(err, e.DBError)
	}
	return nil
}

func DeleteEnv(tx *db.Session, id models.Id) e.Error {
	if _, err := tx.Where("id = ?", id).Delete(&models.Env{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete env error: %v", err))
	}
	return nil
}

func GetEnvById(tx *db.Session, id models.Id) (*models.Env, e.Error) {
	o := models.Env{}
	if err := tx.Model(models.Env{}).Where("id = ?", id).First(&o); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.EnvNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &o, nil
}

func GetEnvByName(tx *db.Session, orgId models.Id, projectId models.Id, name string) (*models.Env, e.Error) {
	o := models.Env{}
	if err := tx.Model(models.Env{}).
		Where("org_id = ? AND project_id = ? AND name = ?", orgId, projectId, name).First(&o); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.EnvNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &o, nil
}

func IsTplAssociationCurrentProject(c *ctx.ServiceContext, tplId models.Id) e.Error {
	if ok, err := c.DB().Model(&models.ProjectTemplate{}).Where("template_id = ?", tplId).Where("project_id = ?", c.ProjectId).Exists(); err != nil {
		return e.New(e.DBError, err)
	} else if !ok {
		return e.New(e.TemplateNotAssociationCurrentProject, fmt.Errorf("the passed tplId is not associated with the current project and cannot create an environment"))
	}
	return nil
}

func QueryEnvDetail(dbSess *db.Session, orgId, projectId models.Id) *db.Session {
	query := dbSess.Where("iac_env.org_id = ? AND iac_env.project_id = ?", orgId, projectId)
	query = query.Model(&models.Env{}).LazySelectAppend("iac_env.*")

	// 模板名称
	query = query.Joins("left join iac_template as t on t.id = iac_env.tpl_id").
		LazySelectAppend("t.name as template_name")
	// 创建人姓名
	query = query.Joins("left join iac_user as u on u.id = iac_env.creator_id").
		LazySelectAppend("u.name as creator")
	// 资源数量统计
	query = query.Joins("left join (select count(*) as resource_count, task_id from iac_resource group by task_id) as r on r.task_id = iac_env.last_res_task_id").
		LazySelectAppend("r.resource_count")
	// 密钥名称
	query = query.Joins("left join iac_key as k on k.id = iac_env.key_id").
		LazySelectAppend("k.name as key_name")
	// 资源是否发生漂移
	query = query.Joins("LEFT JOIN (" +
		"  SELECT iac_resource.task_id FROM iac_resource_drift " +
		"INNER JOIN iac_resource ON iac_resource.id = iac_resource_drift.res_id GROUP BY iac_resource.task_id" +
		") AS rd ON rd.task_id = iac_env.last_res_task_id").
		LazySelectAppend("!ISNULL(rd.task_id) AS is_drift")
	query = query.Joins("left join iac_scan_task on iac_env.last_scan_task_id = iac_scan_task.id").
		LazySelectAppend("iac_scan_task.policy_status as policy_status")

	// 账单数据
	filter := dbSess.Table("iac_bill as b").
		Where("b.org_id = ? and b.project_id = ?", orgId, projectId).
		Where("b.cycle = ?", time.Now().Format("2006-01")).
		Group("b.env_id").
		Select("b.env_id,sum(b.pretax_amount) as month_cost")
	query = query.Joins("left join (?) as b on b.env_id = iac_env.id", filter.Expr()).LazySelectAppend("b.month_cost")

	return query
}

func GetEnvDetailById(query *db.Session, id models.Id) (*models.EnvDetail, e.Error) {
	d := models.EnvDetail{}
	if err := query.Where("iac_env.id = ?", id).First(&d); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.EnvNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &d, nil
}

func GetEnvByTplId(tx *db.Session, tplId models.Id) ([]models.Env, error) {
	env := make([]models.Env, 0)
	if err := tx.Where("tpl_id = ?", tplId).Find(&env); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return env, nil
}

func QueryActiveEnv(query *db.Session) *db.Session {
	return query.Model(&models.Env{}).Where("status in (?,?) OR deploying = ?",
		models.EnvStatusActive, models.EnvStatusFailed, true)
}

func QueryDeploySucessEnv(query *db.Session) *db.Session {
	return query.Model(&models.Env{}).Where("status = ?", models.EnvStatusActive)
}

func QueryEnv(query *db.Session) *db.Session {
	return query.Model(&models.Env{})
}

// ChangeEnvStatusWithTaskAndStep 基于任务和步骤的状态更新环境状态
// nolint:cyclop
func ChangeEnvStatusWithTaskAndStep(tx *db.Session, id models.Id, task *models.Task, step *models.TaskStep) e.Error {
	var (
		envStatus     = ""
		envTaskStatus = ""
		isDeploying   = false
	)

	// 不修改环境数据的任务也不会影响环境状态
	if !task.IsEffectTask() {
		return nil
	}

	if task.Started() && !task.Exited() {
		envTaskStatus = task.Status
		isDeploying = true
	} else if task.Exited() {
		switch task.Status {
		case models.TaskRejected:
			// 任务驳回，环境状态不变
			envStatus = ""
		case models.TaskFailed:
			envStatus = models.EnvStatusFailed
		case models.TaskAborted:
			var err error
			envStatus, err = getEnvStatusOnTaskAborted(tx, task.Id)
			if err != nil {
				return e.New(e.InternalError, errors.Wrap(err, "getEnvStatusOnTaskAborted"))
			}
		case models.TaskComplete:
			if task.Type == models.TaskTypeApply {
				envStatus = models.EnvStatusActive
			} else if task.Type == models.TaskTypeDestroy {
				envStatus = models.EnvStatusDestroyed
			}
		default:
			return e.New(e.InternalError, fmt.Errorf("unknown task status: %v", task.Status))
		}
	} else { // pending
		// 任务进入 pending 状态不修改环境状态， 因为任务 pending 时可能同一个环境的其他任务正在执行
		// (实际目前任务创建后即进入 pending 状态，并不触发 change status 调用链)
		return nil
	}

	logger := logs.Get().WithField("envId", id)
	attrs := models.Attrs{
		"task_status": envTaskStatus,
		"deploying":   isDeploying,
	}
	if envStatus != "" {
		logger.Infof("change env to '%v'", envStatus)
		attrs["status"] = envStatus
	}
	_, err := tx.Model(&models.Env{}).Where("id = ?", id).UpdateAttrs(attrs)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return e.New(e.EnvNotExists)
		}
		return e.New(e.DBError, err)
	}
	return nil
}

// 当任务被中止时需要根据当前执行哪此步骤来判断环境应该置为什么状态
func getEnvStatusOnTaskAborted(db *db.Session, taskId models.Id) (string, error) {
	steps, err := GetTaskSteps(db, taskId)
	if err != nil {
		return "", errors.Wrap(err, "get task steps")
	}

	for _, s := range steps {
		// 如果执行了 apply 步骤则环境变为 failed 状态
		if (s.Type == models.TaskStepApply || s.Type == models.TaskStepDestroy) && s.IsStarted() {
			return models.EnvStatusFailed, nil
		}
	}
	// 否则，环境状态保持不变
	return "", nil
}

var (
	ttlMap = map[string]string{
		"1d":  "24h",
		"3d":  "72h",
		"1w":  "168h",
		"15d": "360h",
		"30d": "720h",
	}
)

func ParseTTL(ttl string) (time.Duration, error) {
	ds, ok := ttlMap[ttl]
	if ok {
		return time.ParseDuration(ds)
	}
	// map 中不存在则尝试直接解析
	t, err := time.ParseDuration(ttl)
	if err != nil {
		return t, fmt.Errorf("invalid duration: %v", ttl)
	}
	return t, nil
}

func GetEnvLastScanTask(sess *db.Session, envId models.Id) (*models.ScanTask, error) {
	task := models.ScanTask{}
	scanTaskIdQuery := sess.Model(&models.Env{}).Where("id = ?", envId).Select("last_scan_task_id")
	err := sess.Model(&models.ScanTask{}).Where("id = (?)", scanTaskIdQuery.Expr()).First(&task)
	return &task, err
}

func GetEnvResourceCount(sess *db.Session, envId models.Id) (int, e.Error) {
	lastResTaskQuery := sess.Model(&models.Env{}).Where("id = ?", envId).Select("last_res_task_id")
	count, err := sess.Model(&models.Resource{}).Where("task_id = (?)", lastResTaskQuery.Expr()).Count()
	if err != nil {
		return 0, e.AutoNew(err, e.DBError)
	}
	return int(count), nil
}

func GetDefaultRunner() (string, e.Error) {
	runners, err := RunnerSearch()
	if err != nil {
		return "", err
	}
	if len(runners) > 0 {
		return runners[0].ID, nil
	}
	return "", e.New(e.ConsulConnError, fmt.Errorf("runner list is null"))
}

func matchVar(v forms.SampleVariables, value models.Variable) bool {
	// 对于第三方调用api创建的环境来说，当前作用域是无变量的，sampleVariables中的变量一种是继承性下来的、另一种是新建的
	// 这里需要判断变量如果修改了就在当前作用域创建一个变量
	// 比较变量名是否相同，相同的变量比较变量的值是否发生变化, 发生变化则创建
	if (v.Name == value.Name && value.Type == consts.VarTypeEnv) ||
		(v.Name == fmt.Sprintf("TF_VAR_%s", value.Name) && value.Type == consts.VarTypeTerraform) {
		return true
	}

	return false
}

func GetRunnerByTags(tags []string) (string, e.Error) {
	runners, err := RunnerSearch()
	if err != nil {
		return "", err
	}

	validRunners := make([]*api.AgentService, 0)
	for _, runner := range runners {
		if utils.ListContains(runner.Tags, tags) {
			validRunners = append(validRunners, runner)
		}
	}

	if len(validRunners) > 0 {
		rand.Seed(time.Now().Unix())
		return validRunners[rand.Intn(len(validRunners))].ID, nil //nolint:gosec
	}

	return "", e.New(e.ConsulConnError, fmt.Errorf("runner list with tags is null"))
}

func GetAvailableRunnerIdByStr(runnerId string, runnerTags string) (string, e.Error) {
	tags := make([]string, 0)
	if runnerTags != "" {
		tags = strings.Split(runnerTags, ",")
	}
	return GetAvailableRunnerId(runnerId, tags)
}

func GetAvailableRunnerId(runnerId string, runnerTags []string) (string, e.Error) {
	if runnerId != "" {
		return runnerId, nil
	}

	if len(runnerTags) > 0 {
		return GetRunnerByTags(runnerTags)
	}
	return GetDefaultRunner()
}

func varNewAppend(resp []forms.Variable, name, value, varType string, sensitive bool) []forms.Variable {
	resp = append(resp, forms.Variable{
		Scope:     consts.ScopeEnv,
		Type:      varType,
		Name:      name,
		Value:     value,
		Sensitive: sensitive,
	})
	return resp
}

func GetSampleValidVariables(tx *db.Session, orgId, projectId, tplId, envId models.Id, sampleVariables []forms.SampleVariables) ([]forms.Variable, e.Error) {
	resp := make([]forms.Variable, 0)
	vars, err, _ := GetValidVariables(tx, consts.ScopeEnv, orgId, projectId, tplId, envId, true)
	if err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("get vairables error: %v", err))
	}
	for _, v := range sampleVariables {
		isNewVaild := true
		// 如果vars为空，则需要将sampleVariables所有的变量理解为新增变量
		if len(vars) == 0 {
			resp = varNewAppend(resp, v.Name, v.Value, consts.VarTypeEnv, v.Sensitive)
			continue
		}

		for key, value := range vars {
			// 如果匹配到了就不在继续匹配
			if matchVar(v, value) {
				// 匹配到了，不管值是否相同都不需要新建变量
				isNewVaild = false
				if v.Value != value.Value {
					resp = varNewAppend(resp, vars[key].Name, v.Value, vars[key].Type, v.Sensitive)
				}
				break
			}
		}

		// 这部分变量是新增的 需要新建
		if isNewVaild {
			resp = varNewAppend(resp, v.Name, v.Value, consts.VarTypeEnv, v.Sensitive)
		}
	}

	return resp, nil
}

// CheckoutAutoApproval 配置漂移自动执行apply、commit自动部署apply是否配置自动审批
func CheckoutAutoApproval(autoApproval, autoDrift bool, triggers []string) bool {
	if autoApproval {
		return true
	}
	// 漂移自动执行apply检测，当勾选漂移自动检测时自动审批同时勾选
	if autoDrift {
		return false
	}

	// 配置commit自动apply时，必须勾选自动审批
	for _, v := range triggers {
		if v == consts.EnvTriggerCommit {
			return false
		}
	}

	return true
}

func CheckEnvTags(tags string) e.Error {
	parts := strings.Split(tags, ",")

	if len(parts) > consts.EnvMaxTagNum {
		return e.New(e.EnvTagNumLimited)
	}

	for _, t := range parts {
		if utf8.RuneCountInString(t) > consts.EnvMaxTagLength {
			return e.New(e.EnvTagLengthLimited)
		}
	}
	return nil
}

func EnvLock(dbSess *db.Session, id models.Id) e.Error {
	if _, err := dbSess.Model(models.Env{}).
		Where("id =?", id).
		UpdateColumn("locked", true); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

func EnvUnLocked(dbSess *db.Session, id models.Id) e.Error {
	if _, err := dbSess.Model(models.Env{}).
		Where("id =?", id).
		UpdateColumn("locked", false); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

// EnvCostTypeStat 费用类型统计
func EnvCostTypeStat(tx *db.Session, id models.Id) ([]resps.EnvCostTypeStatResp, e.Error) {
	/* sample sql:
	select
		t.res_type as res_type,
		SUM(pretax_amount) as amount
	from
		iac_bill
	join (SELECT DISTINCT res_id, type as res_type from iac_resource where iac_resource.env_id  = 'env-c870jh4bh95lubaf3mf0') as t ON
		iac_bill.instance_id = t.res_id
	where iac_bill.cycle = DATE_FORMAT(CURDATE(), "%Y-%m")
	group by
		t.res_type
	*/

	subQuery := tx.Model(&models.Resource{}).Select(`DISTINCT(res_id), iac_resource.type as res_type`)
	subQuery = subQuery.Where(`iac_resource.env_id  = ?`, id)

	query := tx.Model(&models.Bill{}).Select(`t.res_type as res_type, SUM(pretax_amount) as amount`)
	query = query.Joins(`JOIN (?) as t ON iac_bill.instance_id = t.res_id`, subQuery.Expr())

	query = query.Where(`iac_bill.cycle = DATE_FORMAT(CURDATE(), "%Y-%m")`)
	query = query.Group("t.res_type")

	var results []resps.EnvCostTypeStatResp
	if err := query.Find(&results); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return results, nil
}

// EnvCostTrendStat 费用趋势统计
func EnvCostTrendStat(tx *db.Session, id models.Id, months int) ([]resps.EnvCostTrendStatResp, e.Error) {
	/* sample sql:
	select
		iac_bill.cycle as date,
		SUM(pretax_amount) as amount
	from
		iac_bill
	where
		iac_bill.instance_id IN (SELECT DISTINCT  res_id from iac_resource where iac_resource.env_id  = 'env-c870jh4bh95lubaf3mf0')
		AND iac_bill.cycle > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 6 MONTH), "%Y-%m")
	group by
		iac_bill.cycle
	order by
		iac_bill.cycle asc
	*/

	subQuery := tx.Model(&models.Resource{}).Select(`DISTINCT(res_id)`).Where(`iac_resource.env_id  = ?`, id)

	query := tx.Model(&models.Bill{}).Select(`iac_bill.cycle as date, SUM(pretax_amount) as amount`)

	query = query.Where(`iac_bill.instance_id IN (?)`, subQuery.Expr())
	query = query.Where(`iac_bill.cycle > DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL ? MONTH), "%Y-%m")`, months)

	query = query.Group("iac_bill.cycle")
	query = query.Order("iac_bill.cycle asc")

	var results []resps.EnvCostTrendStatResp
	if err := query.Find(&results); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return completeEnvCostTrendData(results, months)
}

// completeEnvCostTrendData 补充缺失的月份数据
func completeEnvCostTrendData(results []resps.EnvCostTrendStatResp, months int) ([]resps.EnvCostTrendStatResp, e.Error) {
	if len(results) == 0 {
		return results, nil
	}

	var allResults = make([]resps.EnvCostTrendStatResp, 0)
	// golang AddDate 减一个月相当于减30天，包含2月份需注意
	now, _ := time.Parse("2006-01", time.Now().Format("2006-01"))
	startMonth := now.AddDate(0, -1*(months-1), 0)

	// 第一个日期加入
	for _, result := range results {
		// 补充缺失的日期
		for {
			startMonthStr := startMonth.Format("2006-01")
			// 日期连续时
			if startMonthStr == result.Date {
				allResults = append(allResults, result)
				startMonth = startMonth.AddDate(0, 1, 0)
				break
			}

			// 日期不连续时，补充0
			allResults = append(allResults, resps.EnvCostTrendStatResp{
				Date:   startMonth.Format("2006-01"),
				Amount: 0.0,
			})
			startMonth = startMonth.AddDate(0, 1, 0)
		}
	}

	return allResults, nil
}

type RawEnvCostDetail struct {
	ResType      string          `json:"resType"`
	Attrs        models.ResAttrs `json:"attrs"`
	Address      string          `json:"address"`
	InstanceId   string          `json:"instanceId"` // 实例id
	CurMonthCost float32         `json:"curMonthCost"`
	TotalCost    float32         `json:"totalCost"`
}

// EnvCostList 费用列表
func EnvCostList(tx *db.Session, id models.Id) ([]RawEnvCostDetail, e.Error) {
	mCurMonth, err := monthEnvCostList(tx, id, true)
	if err != nil {
		return nil, err
	}

	mOtherMonth, err := monthEnvCostList(tx, id, false)
	if err != nil {
		return nil, err
	}

	// 合并当前月和其他月份
	for k, v := range mOtherMonth {
		if _, ok := mCurMonth[k]; !ok {
			mCurMonth[k] = v
			mCurMonth[k].CurMonthCost = 0
		}
	}

	mTotal, err := totalEnvCostListByInstanceId(tx, id)
	if err != nil {
		return nil, err
	}

	// 合并 当前月费用 和 总体费用 的数据
	for k, v := range mTotal {
		if _, ok := mCurMonth[k]; ok {
			mCurMonth[k].TotalCost = v
		}
	}

	var results = make([]RawEnvCostDetail, 0)
	for _, v := range mCurMonth {
		results = append(results, *v)
	}

	return results, nil
}

func monthEnvCostList(tx *db.Session, id models.Id, isCurMonth bool) (map[string]*RawEnvCostDetail, e.Error) {
	/* sample sql:
	select
		iac_resource.attrs as attrs,
		iac_resource.address as address,
		iac_resource.type as res_type,
		iac_bill.instance_id as instance_id,
		pretax_amount as cur_month_cost
	from
		iac_resource
	JOIN iac_bill ON
		iac_bill.instance_id = iac_resource.res_id
	JOIN iac_env ON
		iac_env.id = iac_resource.env_id
		and iac_resource.task_id = iac_env.last_res_task_id
	where
		iac_resource.env_id  = 'env-c8u10aosm56kh90t588g'
		and iac_resource.address NOT LIKE 'data.%'
		and iac_bill.cycle = DATE_FORMAT(CURDATE(), "%Y-%m")
	*/

	query := tx.Model(&models.Resource{}).Select(`iac_resource.attrs as attrs, iac_resource.address as address, iac_resource.type as res_type, iac_bill.instance_id as instance_id, pretax_amount as cur_month_cost`)
	query = query.Joins(`JOIN iac_bill ON iac_bill.instance_id = iac_resource.res_id`)
	query = query.Joins(`JOIN iac_env ON iac_env.id = iac_resource.env_id and iac_resource.task_id = iac_env.last_res_task_id`)

	query = query.Where(`iac_resource.env_id = ?`, id)
	query = query.Where(`iac_resource.address NOT LIKE 'data.%'`)
	if isCurMonth {
		query = query.Where(`iac_bill.cycle = DATE_FORMAT(CURDATE(), "%Y-%m")`)
	} else {
		query = query.Where(`iac_bill.cycle != DATE_FORMAT(CURDATE(), "%Y-%m")`)
	}

	var results []RawEnvCostDetail
	if err := query.Find(&results); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	var m = make(map[string]*RawEnvCostDetail)
	for i, data := range results {
		m[data.InstanceId] = &results[i]
	}

	return m, nil
}

func totalEnvCostListByInstanceId(tx *db.Session, id models.Id) (map[string]float32, e.Error) {
	/* sample sql:
	select
		instance_id,
		SUM(pretax_amount) as total_cost
	from
		iac_bill
	where env_id = 'env-c8u10aosm56kh90t588g'
	group by
		instance_id
	*/

	query := tx.Model(&models.Bill{}).Select(`instance_id, SUM(pretax_amount) as total_cost`)

	query = query.Where(`env_id = ?`, id)
	query = query.Group("instance_id")

	var results []struct {
		InstanceId string
		TotalCost  float32
	}
	if err := query.Find(&results); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	var m = make(map[string]float32)
	for _, data := range results {
		m[data.InstanceId] = data.TotalCost
	}

	return m, nil
}

func FilterEnvStatus(query *db.Session, status string, deploying *bool) (*db.Session, e.Error) {
	if deploying != nil {
		query = query.Where("iac_env.deploying = ?", deploying)
	}

	if status == "" {
		return query, nil
	}

	q := db.Get()

	for _, v := range strings.Split(status, ",") {
		if utils.InArrayStr(models.EnvStatus, v) {
			if status == models.EnvStatusInactive {
				q = q.Or("(iac_env.status = ? or iac_env.status = ?) and iac_env.deploying = 0", v, models.EnvStatusDestroyed)
			} else {
				q = q.Or("iac_env.status = ? and iac_env.deploying = 0", v)
			}
		} else if utils.InArrayStr(models.EnvTaskStatus, v) {
			q = q.Or("iac_env.task_status = ? and iac_env.deploying = 1", v)
		} else {
			return nil, e.New(e.BadParam, http.StatusBadRequest)
		}
	}

	return query.Where(q.Expr()), nil
}

func FilterEnvArchiveStatus(query *db.Session, archiveQ string) (*db.Session, e.Error) {
	switch archiveQ {
	case "":
		// 默认返回未归档环境
		return query.Where("iac_env.archived = 0"), nil
	case "all":
		return query, nil
	case "true":
		// 已归档
		return query.Where("iac_env.archived = 1"), nil
	case "false":
		// 未归档
		return query.Where("iac_env.archived = 0"), nil
	default:
		return nil, e.New(e.BadParam, http.StatusBadRequest)
	}
}

func FilterEnvUpdatedTime(query *db.Session, startTime, endTime *time.Time) *db.Session {
	if startTime == nil || endTime == nil {
		return query
	}

	return query.Where("iac_env.updated_at >= ? and iac_env.updated_at <= ?", startTime, endTime)
}
