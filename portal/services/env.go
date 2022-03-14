// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"fmt"
	"time"
	"strings"
	"unicode/utf8"

	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
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
			return nil, e.New(e.EnvAlreadyExists, err)
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

func QueryEnvDetail(query *db.Session) *db.Session {
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
		"    INNER JOIN iac_resource ON iac_resource.id = iac_resource_drift.res_id GROUP BY iac_resource.task_id" +
		") AS rd ON rd.task_id = iac_env.last_res_task_id").
		LazySelectAppend("!ISNULL(rd.task_id) AS is_drift")
	query = query.Joins("left join iac_scan_task on iac_env.last_scan_task_id = iac_scan_task.id").
		LazySelectAppend("iac_scan_task.policy_status as policy_status")

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
	return query.Model(&models.Env{}).Where("status != ? OR deploying = ?", models.EnvStatusInactive, true)
}

func QueryDeploySucessEnv(query *db.Session) *db.Session {
	return query.Model(&models.Env{}).Where("status = ?", models.EnvStatusActive)
}

func QueryEnv(query *db.Session) *db.Session {
	return query.Model(&models.Env{})
}

// ChangeEnvStatusWithTaskAndStep 基于任务和步骤的状态更新环境状态
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

	if task.Exited() {
		switch task.Status {
		case models.TaskRejected:
			// 任务驳回，环境状态不变
			break
		case models.TaskFailed:
			envStatus = models.EnvStatusFailed
		case models.TaskComplete:
			if task.Type == models.TaskTypeApply {
				envStatus = models.EnvStatusActive
			} else if task.Type == models.TaskTypeDestroy {
				envStatus = models.EnvStatusInactive
			}
		default:
			return e.New(e.InternalError, fmt.Errorf("unknown exited task status: %v", task.Status))
		}
	} else if task.Started() {
		envTaskStatus = task.Status
		isDeploying = true
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

func varNewAppend(resp []forms.Variable, name, value, varType string) []forms.Variable {
	resp = append(resp, forms.Variable{
		Scope: consts.ScopeEnv,
		Type:  varType,
		Name:  name,
		Value: value,
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
			resp = varNewAppend(resp, v.Name, v.Value, consts.VarTypeEnv)
			continue
		}

		for key, value := range vars {
			// 如果匹配到了就不在继续匹配
			if matchVar(v, value) {
				if v.Value != value.Value {
					isNewVaild = false
					resp = varNewAppend(resp, vars[key].Name, v.Value, vars[key].Type)
				}
				break
			}
		}

		// 这部分变量是新增的 需要新建
		if isNewVaild{
			resp = varNewAppend(resp, v.Name, v.Value, consts.VarTypeEnv)
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
