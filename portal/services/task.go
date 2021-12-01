// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/common"
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services/logstorage"
	"cloudiac/portal/services/notificationrc"
	"cloudiac/portal/services/vcsrv"
	"cloudiac/runner"
	"cloudiac/utils"
	"cloudiac/utils/kafka"
	"cloudiac/utils/logs"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

func GetTask(dbSess *db.Session, id models.Id) (*models.Task, e.Error) {
	task := models.Task{}
	err := dbSess.Where("id = ?", id).First(&task)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TaskNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &task, nil
}

func DeleteTaskStep(tx *db.Session, taskId models.Id) e.Error {
	step := models.TaskStep{}
	_, err := tx.Where("task_id = ?", taskId).Delete(&step)
	if err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

func DeleteTask(tx *db.Session, taskId models.Id) e.Error {
	step := models.Task{}
	_, err := tx.Where("id = ?", taskId).Delete(&step)
	if err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

func DeleteHistoryCronTask(tx *db.Session) e.Error {
	task := make([]models.Task, 0)
	err := tx.Where("unix_timestamp(now()) - unix_timestamp(end_at) > ? and is_drift_task = ? and type = ?",
		86400*7, true, models.TaskTypePlan).Find(&task)
	if err != nil {
		return e.New(e.DBError, fmt.Errorf("delete task error: %v", err))
	}
	// 删除任务以及任务相关的step
	if len(task) > 0 {
		for _, v := range task {
			if er1 := DeleteTask(tx, v.Id); er1 != nil {
				return er1
			}
			if er1 := DeleteTaskStep(tx, v.Id); er1 != nil {
				return er1
			}
		}
	}
	return nil
}

func GetResourceIdByAddressAndTaskId(sess *db.Session, address string, lastResTaskId models.Id) (*models.Resource, e.Error) {
	res := models.Resource{}
	if err := sess.Where("address = ? and task_id = ?", address, lastResTaskId).First(&res); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return &res, nil
}

func CloneTask(tx *db.Session, pt models.Task, env *models.Env) (*models.Task, e.Error) {
	logger := logs.Get().WithField("func", "CreateTask")
	var err error
	pt.Id = models.NewId("run")
	logger = logger.WithField("taskId", pt.Id)
	tpl, err := GetTemplateById(tx, pt.TplId)
	if err != nil {
		return nil, e.New(e.InternalError, err)
	}
	pt.RepoAddr, _, err = GetTaskRepoAddrAndCommitId(tx, tpl, pt.Revision)
	// 克隆任务需要重置部分任务参数
	var cronTaskType string
	if env.AutoRepairDrift {
		cronTaskType = models.TaskTypeApply
	} else {
		cronTaskType = models.TaskTypePlan
	}
	pt.Name = models.Task{}.GetTaskNameByType(cronTaskType)
	pt.Type = cronTaskType
	pt.Status = models.TaskPending
	pt.CurrStep = 0
	pt.CreatorId = consts.SysUserId
	pt.Name = common.CronDriftTaskName
	pt.Result = models.TaskResult{}
	pt.CreatedAt = models.Time{}
	pt.UpdatedAt = models.Time{}
	pt.StartAt = &models.Time{}
	pt.EndAt = &models.Time{}
	pt.ContainerId = ""
	if err != nil {
		return nil, e.New(e.InternalError, err)
	}
	return doCreateTask(tx, pt, tpl, env)

}

func CreateTask(tx *db.Session, tpl *models.Template, env *models.Env, pt models.Task) (*models.Task, e.Error) {
	logger := logs.Get().WithField("func", "CreateTask")

	var (
		err      error
		firstVal = utils.FirstValueStr
	)
	task := models.Task{
		// 以下为需要外部传入的属性
		Name:            pt.Name,
		Targets:         pt.Targets,
		CreatorId:       pt.CreatorId,
		Variables:       pt.Variables,
		AutoApprove:     pt.AutoApprove,
		KeyId:           models.Id(firstVal(string(pt.KeyId), string(env.KeyId))),
		ExtraData:       pt.ExtraData,
		Revision:        firstVal(pt.Revision, env.Revision, tpl.RepoRevision),
		StopOnViolation: pt.StopOnViolation,

		RetryDelay:  utils.FirstValueInt(pt.RetryDelay, env.RetryDelay),
		RetryNumber: utils.FirstValueInt(pt.RetryNumber, env.RetryNumber),
		RetryAble:   env.RetryAble,

		OrgId:     env.OrgId,
		ProjectId: env.ProjectId,
		TplId:     env.TplId,
		EnvId:     env.Id,
		StatePath: env.StatePath,

		Workdir:   tpl.Workdir,
		TfVersion: tpl.TfVersion,

		// 以下值直接使用环境的配置(不继承模板的配置)
		Playbook:     env.Playbook,
		TfVarsFile:   env.TfVarsFile,
		PlayVarsFile: env.PlayVarsFile,

		BaseTask: models.BaseTask{
			Type:        pt.Type,
			Pipeline:    pt.Pipeline,
			StepTimeout: utils.FirstValueInt(pt.StepTimeout, common.DefaultTaskStepTimeout),
			RunnerId:    firstVal(pt.RunnerId, env.RunnerId),

			Status:   models.TaskPending,
			Message:  "",
			CurrStep: 0,
		},
		Callback: pt.Callback,
	}

	task.Id = models.NewId("run")
	logger = logger.WithField("taskId", task.Id)

	task.RepoAddr, task.CommitId, err = GetTaskRepoAddrAndCommitId(tx, tpl, task.Revision)
	if err != nil {
		return nil, e.New(e.InternalError, err)
	}
	return doCreateTask(tx, task, tpl, env)

}

func doCreateTask(tx *db.Session, task models.Task, tpl *models.Template, env *models.Env) (*models.Task, e.Error) {
	// pipeline 内容可以从外部传入，如果没有传则尝试读取云模板目录下的文件
	var err error
	logger := logs.Get().WithField("func", "doCreateTask")
	{ // 参数检查
		if task.Playbook != "" && task.KeyId == "" {
			return nil, e.New(e.BadParam, fmt.Errorf("'keyId' is required to run playbook"))
		}
		if task.RepoAddr == "" {
			return nil, e.New(e.BadParam, fmt.Errorf("'repoAddr' is required"))
		}
		if task.CommitId == "" {
			return nil, e.New(e.BadParam, fmt.Errorf("'commitId' is required"))
		}
		if task.RunnerId == "" {
			return nil, e.New(e.BadParam, fmt.Errorf("'runnerId' is required"))
		}
	}
	if task.Pipeline == "" {
		task.Pipeline, err = GetTplPipeline(tx, tpl.Id, task.Revision, task.Workdir)
		if err != nil {
			return nil, e.AutoNew(err, e.InvalidPipeline)
		}
	}

	pipeline, err := DecodePipeline(task.Pipeline)
	if err != nil {
		return nil, e.New(e.InvalidPipeline, err)
	}

	task.Flow = GetTaskFlowWithPipeline(pipeline, task.Type)
	steps := make([]models.TaskStep, 0)
	stepIndex := 0
	for i, pipelineStep := range task.Flow.Steps {
		if pipelineStep.Type == models.TaskStepPlay && task.Playbook == "" {
			logger.WithField("step", fmt.Sprintf("%d(%s)", i, pipelineStep.Type)).
				Infof("not have playbook, skip this step")
			continue
		} else if pipelineStep.Type == models.TaskStepOpaScan {
			// 如果环境扫描未启用，则跳过扫描步骤
			if enabled, _ := IsEnvEnabledScan(tx, task.EnvId); !enabled {
				continue
			}
		}

		if len(task.Targets) != 0 && IsTerraformStep(pipelineStep.Type) {
			if pipelineStep.Type != models.TaskStepInit {
				for _, t := range task.Targets {
					pipelineStep.Args = append(pipelineStep.Args, fmt.Sprintf("-target=%s", t))
				}
			}
		}

		if pipelineStep.Type == models.TaskStepOpaScan {
			// 对于包含扫描的任务，创建一个对应的 scanTask 作为扫描任务记录，便于后期扫描状态的查询
			scanTask := CreateMirrorScanTask(&task)
			if _, err := tx.Save(scanTask); err != nil {
				return nil, e.New(e.DBError, err)
			}
			env.LastScanTaskId = scanTask.Id
			if _, err := tx.Save(env); err != nil {
				return nil, e.New(e.DBError, errors.Wrapf(err, "update env scan task id"))
			}
		}

		taskStep := newTaskStep(tx, task, pipelineStep, stepIndex)

		steps = append(steps, *taskStep)
		stepIndex += 1
	}

	if len(steps) == 0 {
		return nil, e.New(e.TaskNotHaveStep, fmt.Errorf("task have no steps"))
	}

	if err = tx.Insert(&task); err != nil {
		return nil, e.New(e.DBError, errors.Wrapf(err, "save task"))
	}

	for i := range steps {
		if i+1 < len(steps) {
			steps[i].NextStep = steps[i+1].Id
		}
		if err = tx.Insert(&steps[i]); err != nil {
			return nil, e.New(e.DBError, errors.Wrapf(err, "save task step"))
		}
	}
	return &task, nil
}

func GetTaskRepoAddrAndCommitId(tx *db.Session, tpl *models.Template, revision string) (repoAddr, commitId string, err e.Error) {
	var (
		u         *url.URL
		er        error
		repoToken = tpl.RepoToken
	)

	repoAddr = tpl.RepoAddr
	if tpl.VcsId == "" { // 用户直接填写的 repo 地址
		commitId = revision
	} else {
		var (
			vcs  *models.Vcs
			repo vcsrv.RepoIface
		)
		if vcs, err = QueryVcsByVcsId(tpl.VcsId, tx); err != nil {
			if e.IsRecordNotFound(err) {
				return "", "", e.New(e.VcsNotExists, err)
			}
			return "", "", e.New(e.DBError, err)
		}

		repo, er = vcsrv.GetRepo(vcs, tpl.RepoId)
		if er != nil {
			return "", "", e.New(e.VcsError, er)
		}

		commitId, er = repo.BranchCommitId(revision)
		if er != nil {
			return "", "", e.New(e.VcsError, er)
		}

		if repoAddr == "" {
			// 如果模板中没有记录 repoAddr，则动态获取
			repoAddr, er = vcsrv.GetRepoAddress(repo)
			if er != nil {
				return "", "", e.New(e.VcsError, er)
			}
		} else if !strings.Contains(repoAddr, "://") {
			// 如果 addr 不是完整路径则添加上 vcs 的 address(这样可以允许保存相对路径到 repoAddr)
			repoAddr = utils.JoinURL(vcs.Address, repoAddr)
		}

		if repoToken == "" {
			token, err := vcs.DecryptToken()
			if err != nil {
				return "", "", e.New(e.VcsError, er)
			}
			repoToken = token
		}
	}

	if repoAddr == "" {
		return "", "", e.New(e.BadParam, fmt.Errorf("repo address is blank"))
	}

	u, er = url.Parse(repoAddr)
	if er != nil {
		return "", "", e.New(e.InternalError, errors.Wrapf(er, "parse url: %v", repoAddr))
	} else if repoToken != "" {
		u.User = url.UserPassword("token", repoToken)
	}
	repoAddr = u.String()

	return repoAddr, commitId, nil
}

func ListPendingCronTask(tx *db.Session, envId models.Id) (bool, e.Error) {
	query := tx.Where("env_id = ? and is_drift_task = true and (status = ? or status= ?)", envId, models.TaskPending, models.TaskRunning)
	exist, err := query.Model(&models.Task{}).Exists()
	if err != nil {
		return exist, e.New(e.DBError, err)
	}
	return exist, nil
}

func GetTaskById(tx *db.Session, id models.Id) (*models.Task, e.Error) {
	o := models.Task{}
	if err := tx.Where("id = ?", id).First(&o); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TaskNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &o, nil
}

func QueryTask(query *db.Session) *db.Session {
	query = query.Model(&models.Task{})
	// 创建人姓名
	query = query.Joins("left join iac_user as u on u.id = iac_task.creator_id").
		LazySelectAppend("u.name as creator,iac_task.*")
	return query
}

var stepStatus2TaskStatusMap = map[string]string{
	// 步骤进入 pending，将任务标识为 running
	// (正常情况下步骤进入 pending 并不会触发 ChangeXXXStatus 调用，只有在步骤通过审批时会走到这个逻辑)
	models.TaskStepPending:   models.TaskRunning,
	models.TaskStepApproving: models.TaskApproving,
	models.TaskStepRejected:  models.TaskRejected,
	models.TaskStepRunning:   models.TaskRunning,
	models.TaskStepFailed:    models.TaskFailed,
	models.TaskStepTimeout:   models.TaskFailed,
	models.TaskStepComplete:  models.TaskComplete,
}

func stepStatus2TaskStatus(s string) string {
	taskStatus, ok := stepStatus2TaskStatusMap[s]
	if !ok {
		panic(fmt.Errorf("unknown task step status %v", s))
	}
	return taskStatus
}

func ChangeTaskStatusWithStep(dbSess *db.Session, task models.Tasker, step *models.TaskStep) e.Error {
	switch t := task.(type) {
	case *models.Task:
		return ChangeTaskStatus(dbSess, t, stepStatus2TaskStatus(step.Status), step.Message)
	case *models.ScanTask:
		return ChangeScanTaskStatusWithStep(dbSess, t, step)
	default:
		panic("invalid task type on change task status with step, task" + task.GetId())
	}
}

// ChangeTaskStatus 修改任务状态(同步修改 StartAt、EndAt 等)，并同步修改 env 状态
func ChangeTaskStatus(dbSess *db.Session, task *models.Task, status, message string) e.Error {
	preStatus := task.Status
	if preStatus == status && message == "" {
		return nil
	}

	task.Status = status
	task.Message = message
	now := models.Time(time.Now())
	if task.StartAt == nil && task.Started() {
		task.StartAt = &now
	}
	if task.EndAt == nil && task.Exited() {
		task.EndAt = &now
	}

	logs.Get().WithField("taskId", task.Id).Infof("change task to '%s'", status)
	if _, err := dbSess.Model(task).Update(task); err != nil {
		return e.AutoNew(err, e.DBError)
	}

	// 回调的消息通知只发送一次, 作业结束后发送通知
	if task.Callback != "" && task.Exited() {
		switch task.Callback {
		case consts.TaskCallbackKafka:
			SendKafkaMessage(dbSess, task, status)
		default:
			logs.Get().Infof("callback type don't support")
		}
	}

	if preStatus != status && !task.IsDriftTask {
		TaskStatusChangeSendMessage(task, status)
	}
	// 如果勾选提交pr自动plan，任务结束时 plan作业结果写入PR评论中
	if task.Exited() && task.Type == common.TaskTypePlan {
		SendVcsComment(dbSess, task, status)
	}

	step, er := GetTaskStep(dbSess, task.Id, task.CurrStep)
	if er != nil {
		return e.AutoNew(er, e.DBError)
	}
	return ChangeEnvStatusWithTaskAndStep(dbSess, task.EnvId, task, step)
}

type TfState struct {
	FormVersion      string        `json:"form_version"`
	TerraformVersion string        `json:"terraform_version"`
	Values           TfStateValues `json:"values"`
}

// TfStateValues  doc: https://www.terraform.io/docs/internals/json-format.html#values-representation
type TfStateValues struct {
	Outputs      map[string]TfStateVariable `json:"outputs"`
	RootModule   TfStateModule              `json:"root_module"`
	ChildModules []TfStateModule            `json:"child_modules,omitempty"`
}

type TfStateModule struct {
	Address      string            `json:"address"`
	Resources    []TfStateResource `json:"resources"`
	ChildModules []TfStateModule   `json:"child_modules,omitempty"`
}

type TfStateVariable struct {
	Value     interface{} `json:"value"`
	Sensitive bool        `json:"sensitive,omitempty"`
}

type TfStateResource struct {
	ProviderName string      `json:"provider_name"`
	Address      string      `json:"address"`
	Mode         string      `json:"mode"` // managed、data
	Type         string      `json:"type"`
	Name         string      `json:"name"`
	Index        interface{} `json:"index"` // index 可以为整型或字符串

	Values map[string]interface{} `json:"values"`
}

func UnmarshalStateJson(bs []byte) (*TfState, error) {
	state := TfState{}
	err := json.Unmarshal(bs, &state)
	return &state, err
}

func traverseStateModule(module *TfStateModule) (rs []*models.Resource) {
	parts := strings.Split(module.Address, ".")
	moduleName := parts[len(parts)-1]
	for _, r := range module.Resources {
		idx := ""
		if r.Index != nil {
			idx = fmt.Sprintf("%v", r.Index)
		}
		rs = append(rs, &models.Resource{
			Provider: r.ProviderName,
			Module:   moduleName,
			Address:  r.Address,
			Type:     r.Type,
			Name:     r.Name,
			Index:    idx,
			Attrs:    r.Values,
		})
	}

	for i := range module.ChildModules {
		rs = append(rs, traverseStateModule(&module.ChildModules[i])...)
	}
	return rs
}

func SaveTaskResources(tx *db.Session, task *models.Task, values TfStateValues, proMap runner.ProviderSensitiveAttrMap) error {

	bq := utils.NewBatchSQL(1024, "INSERT INTO", models.Resource{}.TableName(),
		"id", "org_id", "project_id", "env_id", "task_id",
		"provider", "module", "address", "type", "name", "index", "attrs", "sensitive_keys")

	rs := make([]*models.Resource, 0)
	rs = append(rs, traverseStateModule(&values.RootModule)...)
	for i := range values.ChildModules {
		rs = append(rs, traverseStateModule(&values.ChildModules[i])...)
	}

	for _, r := range rs {
		if len(proMap) > 0 {
			providerKey := strings.Join([]string{r.Provider, r.Type}, "-")
			// 通过provider-type 拼接查找敏感词是否在proMap中
			if _, ok := proMap[providerKey]; ok {
				r.SensitiveKeys = proMap[providerKey]
			}
		}
		err := bq.AddRow(models.NewId("r"), task.OrgId, task.ProjectId, task.EnvId, task.Id,
			r.Provider, r.Module, r.Address, r.Type, r.Name, r.Index, r.Attrs, r.SensitiveKeys)
		if err != nil {
			return err
		}
	}

	for bq.HasNext() {
		sql, args := bq.Next()
		if _, err := tx.Exec(sql, args...); err != nil {
			return err
		}
	}
	return nil
}

func SaveTaskOutputs(dbSess *db.Session, task *models.Task, vars map[string]TfStateVariable) error {
	task.Result.Outputs = make(map[string]interface{}, 0)
	for k, v := range vars {
		task.Result.Outputs[k] = v
	}
	if _, err := dbSess.Model(&models.Task{}).Where("id = ?", task.Id).
		UpdateColumn("result", task.Result); err != nil {
		return err
	}
	return nil
}

// TfPlan doc: https://www.terraform.io/docs/internals/json-format.html#plan-representation
type TfPlan struct {
	FormatVersion string `json:"format_version"`

	ResourceChanges []TfPlanResource `json:"resource_changes"`
}

type TfPlanResource struct {
	Address       string `json:"address"`
	ModuleAddress string `json:"module_address,omitempty"`

	Mode  string `json:"mode"` // managed、data
	Type  string `json:"type"`
	Name  string `json:"name"`
	Index int    `json:"index"`

	Change TfPlanResourceChange `json:"change"`
}

// TfPlanResourceChange doc: https://www.terraform.io/docs/internals/json-format.html#change-representation
type TfPlanResourceChange struct {
	Actions []string    `json:"actions"` // no-op, create, read, update, delete
	Before  interface{} `json:"before"`
	After   interface{} `json:"after"`
}

func UnmarshalPlanJson(bs []byte) (*TfPlan, error) {
	plan := TfPlan{}
	err := json.Unmarshal(bs, &plan)
	return &plan, err
}

type TSResource struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	ModuleName string `json:"module_name"`
	Source     string `json:"source"`
	PlanRoot   string `json:"plan_root"`
	Line       int    `json:"line"`
	Type       string `json:"type"`

	Config map[string]interface{} `json:"config"`

	SkipRules   *bool  `json:"skip_rules"`
	MaxSeverity string `json:"max_severity"`
	MinSeverity string `json:"min_severity"`
}

type TSResources []TSResource

type TfParse map[string]TSResources

func UnmarshalTfParseJson(bs []byte) (*TfParse, error) {
	js := TfParse{}
	err := json.Unmarshal(bs, &js)
	return &js, err
}

type TsResultJson struct {
	Results TsResult `json:"results"`
}

type TsResult struct {
	ScanErrors        []ScanError `json:"scan_errors,omitempty"`
	PassedRules       []Rule      `json:"passed_rules,omitempty"`
	Violations        []Violation `json:"violations"`
	SkippedViolations []Violation `json:"skipped_violations"`
	ScanSummary       ScanSummary `json:"scan_summary"`
}

type ScanError struct {
	IacType   string `json:"iac_type"`
	Directory string `json:"directory"`
	ErrMsg    string `json:"errMsg"`
}

type ScanSummary struct {
	FileFolder        string `json:"file/folder"`
	IacType           string `json:"iac_type"`
	ScannedAt         string `json:"scanned_at"`
	PoliciesValidated int    `json:"policies_validated"`
	ViolatedPolicies  int    `json:"violated_policies"`
	Low               int    `json:"low"`
	Medium            int    `json:"medium"`
	High              int    `json:"high"`
}

type Rule struct {
	RuleName    string `json:"rule_name"`
	Description string `json:"description"`
	RuleId      string `json:"rule_id"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
}

type Violation struct {
	RuleName     string `json:"rule_name"`
	Description  string `json:"description"`
	RuleId       string `json:"rule_id"`
	Severity     string `json:"severity"`
	Category     string `json:"category"`
	Comment      string `json:"skip_comment,omitempty"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	ModuleName   string `json:"module_name,omitempty"`
	File         string `json:"file,omitempty"`
	PlanRoot     string `json:"plan_root,omitempty"`
	Line         int    `json:"line,omitempty"`
	Source       string `json:"source,omitempty"`
}

type TsCount struct {
	Low    int `json:"low"`
	Medium int `json:"medium"`
	High   int `json:"high"`
	Total  int `json:"total"`
}

func UnmarshalTfResultJson(bs []byte) (*TsResultJson, error) {
	js := TsResultJson{}
	err := json.Unmarshal(bs, &js)
	return &js, err
}

func SaveTaskChanges(dbSess *db.Session, task *models.Task, rs []TfPlanResource) error {
	var (
		resAdded     = 0
		resChanged   = 0
		resDestroyed = 0
	)
	for _, r := range rs {
		actions := r.Change.Actions
		switch {
		case utils.SliceEqualStr(actions, []string{"no-op"}),
			utils.SliceEqualStr(actions, []string{"create", "delete"}):
			continue
		case utils.SliceEqualStr(actions, []string{"create"}):
			resAdded += 1
		case utils.SliceEqualStr(actions, []string{"update"}),
			utils.SliceEqualStr(actions, []string{"delete", "create"}):
			resChanged += 1
		case utils.SliceEqualStr(actions, []string{"delete"}):
			resDestroyed += 1
		default:
			logs.Get().WithField("taskId", task.Id).Errorf("unknown change actions: %v", actions)
		}
	}

	task.Result.ResAdded = &resAdded
	task.Result.ResChanged = &resChanged
	task.Result.ResDestroyed = &resDestroyed

	if _, err := dbSess.Model(&models.Task{}).Where("id = ?", task.Id).
		UpdateColumn("result", task.Result); err != nil {
		return err
	}
	return nil
}

func GetTaskStepByStepId(tx *db.Session, stepId models.Id) (*models.TaskStep, error) {
	taskStep := models.TaskStep{}
	err := tx.Where("id = ?", stepId).First(&taskStep)
	if err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TaskStepNotExists)
		}
		return nil, e.New(e.DBError, err)
	}
	return &taskStep, nil
}

func FetchTaskLog(ctx context.Context, task models.Tasker, stepType string, writer io.WriteCloser) (err error) {
	// close 后 read 端会触发 EOF error
	defer writer.Close()

	var (
		steps        []*models.TaskStep
		fetchedSteps = make(map[string]struct{})
	)

	steps, err = GetTaskSteps(db.Get(), task.GetId())
	if err != nil {
		return err
	}

	for {
		for _, s := range steps {
			if stepType != "" && s.Type != stepType {
				continue
			}
			if _, ok := fetchedSteps[s.Id.String()]; ok {
				continue
			}
			if err := fetchTaskStepLog(ctx, task, writer, s.Id); err != nil {
				return err
			}
			fetchedSteps[s.Id.String()] = struct{}{}
		}

		if task.Exited() {
			break
		}

		// 因为有 callback 步骤，所以任务的步骤是会新增的(但只加到末尾)。
		// 我们等待一定时间，确保没有新的步骤了才退出循环
		time.Sleep(consts.DbTaskPollInterval * 2)
		steps, err = GetTaskSteps(db.Get(), task.GetId())
		if err != nil {
			return err
		}
		if len(steps) <= len(fetchedSteps) {
			break
		}
	}

	return nil
}

func FetchTaskStepLog(ctx context.Context, task models.Tasker, writer io.WriteCloser, stepId models.Id) (err error) {
	// close 后 read 端会触发 EOF error
	defer writer.Close()
	return fetchTaskStepLog(ctx, task, writer, stepId)
}

func fetchTaskStepLog(ctx context.Context, task models.Tasker, writer io.Writer, stepId models.Id) (err error) {

	step, err := GetTaskStepByStepId(db.Get(), stepId)
	if err != nil {
		return err
	}

	sleepDuration := consts.DbTaskPollInterval
	storage := logstorage.Get()
	ticker := time.NewTicker(sleepDuration)
	defer ticker.Stop()

	for _, step := range []*models.TaskStep{step} {
		//logger := logs.Get().WithField("step", fmt.Sprintf("%s(%d)", step.Type, step.Index))
		// 任务有可能未开始执行步骤就退出了，所以需要先判断任务是否退出
		for !task.Exited() && !step.IsStarted() {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				var (
					// 先保存一下值，因为查询出错时 task、step 会被赋值为空，无法用于生成 error
					taskId    = task.GetId()
					stepIndex = step.Index
				)
				task, err = GetTask(db.Get(), task.GetId())
				if err != nil {
					return errors.Wrapf(err, "get task '%s'", taskId)
				}
				step, err = GetTaskStep(db.Get(), task.GetId(), step.Index)
				if err != nil {
					return errors.Wrapf(err, "get task step %d", stepIndex)
				}
			}
		}

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if step.IsExited() {
			var content []byte
			if content, err = storage.Read(step.LogPath); err != nil {
				if os.IsNotExist(err) {
					// 当前步骤没有日志文件，继续读下一步骤
					continue
				}
				return err
			} else if _, err = writer.Write(content); err != nil {
				return err
			}
		} else if step.IsStarted() { // running
			for {
				if err = fetchRunnerTaskStepLog(ctx, task.GetRunnerId(), step, writer); err != nil {
					if err == ErrRunnerTaskNotExists && step.StartAt != nil &&
						time.Since(time.Time(*step.StartAt)) < consts.RunnerConnectTimeout*2 {
						// 某些情况下可能步骤被标识为了 running 状态，但调用 runner 执行任务时因为网络等原因导致没有及时启动执行。
						// 所以这里加一个判断, 如果是刚启动的任务会进行重试
						time.Sleep(sleepDuration)
						continue
					}
					return err
				}
				break
			}
		}
	}

	return nil
}

var (
	ErrRunnerTaskNotExists = errors.New("runner task not exists")
)

// 从 runner 获取任务日志，直到任务结束
func fetchRunnerTaskStepLog(ctx context.Context, runnerId string, step *models.TaskStep, writer io.Writer) error {
	logger := logs.Get().WithField("func", "fetchRunnerTaskStepLog").
		WithField("taskId", step.TaskId).
		WithField("step", fmt.Sprintf("%d(%s)", step.Index, step.Type))

	runnerAddr, err := GetRunnerAddress(runnerId)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Add("envId", string(step.EnvId))
	params.Add("taskId", string(step.TaskId))
	params.Add("step", fmt.Sprintf("%d", step.Index))
	wsConn, resp, err := utils.WebsocketDail(runnerAddr, consts.RunnerTaskStepLogFollowURL, params)
	if err != nil {
		if resp != nil {
			if resp.StatusCode == http.StatusNotFound {
				return ErrRunnerTaskNotExists
			}
			respBody, _ := io.ReadAll(resp.Body)
			logger.Warnf("websocket dail error: %s, response: %s", err, respBody)
		}
		return errors.Wrapf(err, "websocket dail: %s/%s", runnerAddr, consts.RunnerTaskStepLogFollowURL)
	}

	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		_ = utils.WebsocketClose(wsConn)
	}()

	for {
		var reader io.Reader

		_, reader, err = wsConn.NextReader()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				logger.Tracef("read message error: %v", err)
				return nil
			} else {
				logger.Warnf("read message error: %v", err)
				return err
			}
		} else {
			select {
			case <-ctx.Done():
				return nil
			default:
			}

			_, err = io.Copy(writer, reader)
			if err != nil {
				if err == io.ErrClosedPipe {
					return nil
				}
				logger.Warnf("io copy error: %v", err)
				return err
			}
		}
	}
}

func TaskStatusChangeSendMessage(task *models.Task, status string) {
	// 非通知类型的状态直接跳过
	if _, ok := consts.TaskStatusToEventType[status]; !ok {
		logs.Get().WithField("taskId", task.Id).Infof("event don't need send message")
		return
	}
	dbSess := db.Get()
	env, _ := GetEnv(dbSess, task.EnvId)
	tpl, _ := GetTemplateById(dbSess, task.TplId)
	project, _ := GetProjectsById(dbSess, task.ProjectId)
	org, _ := GetOrganizationById(dbSess, task.OrgId)
	ns := notificationrc.NewNotificationService(&notificationrc.NotificationOptions{
		OrgId:     task.OrgId,
		ProjectId: task.ProjectId,
		Tpl:       tpl,
		Project:   project,
		Org:       org,
		Env:       env,
		Task:      task,
		EventType: consts.TaskStatusToEventType[status],
	})
	logs.Get().WithField("taskId", task.Id).Infof("new event: %s", ns.EventType)
	ns.SendMessage()
}

// ==================================================================================
// 扫描任务

// ChangeScanTaskStatus 修改扫描任务状态
func ChangeScanTaskStatus(dbSess *db.Session, task *models.ScanTask, status, message string) e.Error {
	if task.Status == status && message == "" {
		return nil
	}

	task.Status = status
	task.Message = message
	now := models.Time(time.Now())
	if task.StartAt == nil && task.Started() {
		task.StartAt = &now
	}
	if task.EndAt == nil && task.Exited() {
		task.EndAt = &now
	}

	logs.Get().WithField("taskId", task.Id).Infof("change scan task to '%s'", status)
	if _, err := dbSess.Model(task).Update(task); err != nil {
		return e.AutoNew(err, e.DBError)
	}

	return nil
}

func ChangeScanTaskStatusWithStep(dbSess *db.Session, task *models.ScanTask, step *models.TaskStep) e.Error {
	taskStatus := stepStatus2TaskStatus(step.Status)
	exitCode := step.ExitCode

	switch taskStatus {
	case common.TaskPending, common.TaskRunning:
		task.PolicyStatus = common.PolicyStatusPending
	case common.TaskComplete:
		task.PolicyStatus = common.PolicyStatusPassed
	case common.TaskFailed:
		if step.Type == common.TaskStepOpaScan && exitCode == common.TaskStepPolicyViolationExitCode {
			task.PolicyStatus = common.PolicyStatusViolated
		} else {
			task.PolicyStatus = common.PolicyStatusFailed
		}
	default: // "approving", "rejected", ...
		panic(fmt.Errorf("invalid scan task status '%s'", taskStatus))
	}
	return ChangeScanTaskStatus(dbSess, task, taskStatus, step.Message)
}

func CreateScanTask(tx *db.Session, tpl *models.Template, env *models.Env, pt models.ScanTask) (*models.ScanTask, e.Error) {
	logger := logs.Get().WithField("func", "CreateScanTask")

	var (
		err error
		er  e.Error
	)
	envRevison := ""

	envId := models.Id("")
	if env != nil { // env != nil 表示为环境扫描任务
		tpl, er = GetTemplateById(tx, env.TplId)
		if er != nil {
			return nil, e.New(er.Code(), err, http.StatusBadRequest)
		}
		envId = env.Id
		envRevison = env.Revision
	}

	task := models.ScanTask{
		// 以下为需要外部传入的属性
		Name:      pt.Name,
		CreatorId: pt.CreatorId,
		Extra:     pt.Extra,
		Revision:  utils.FirstValueStr(pt.Revision, envRevison, tpl.RepoRevision),

		OrgId:     tpl.OrgId,
		TplId:     tpl.Id,
		EnvId:     envId,
		ProjectId: pt.ProjectId,

		Workdir: tpl.Workdir,

		PolicyStatus: common.PolicyStatusPending,

		BaseTask: models.BaseTask{
			Type:        pt.Type,
			StepTimeout: utils.FirstValueInt(pt.StepTimeout, common.DefaultTaskStepTimeout),
			RunnerId:    pt.RunnerId,

			Status:   models.TaskPending,
			Message:  "",
			CurrStep: 0,
		},
	}

	task.Id = models.NewId("run")
	logger = logger.WithField("taskId", task.Id)

	task.RepoAddr, task.CommitId, err = GetTaskRepoAddrAndCommitId(tx, tpl, task.Revision)
	if err != nil {
		return nil, e.New(e.InternalError, err)
	}

	{ // 参数检查
		if task.RepoAddr == "" {
			return nil, e.New(e.BadParam, fmt.Errorf("'repoAddr' is required"))
		}
		if task.CommitId == "" {
			return nil, e.New(e.BadParam, fmt.Errorf("'commitId' is required"))
		}
		if task.RunnerId == "" {
			return nil, e.New(e.BadParam, fmt.Errorf("'runnerId' is required"))
		}
	}

	task.Pipeline, err = GetTplPipeline(tx, tpl.Id, task.Revision, task.Workdir)
	if err != nil {
		return nil, e.AutoNew(err, e.InvalidPipeline)
	}

	pipeline, err := DecodePipeline(task.Pipeline)
	if err != nil {
		return nil, e.New(e.InvalidPipeline, err)
	}

	task.Flow = GetTaskFlowWithPipeline(pipeline, task.Type)
	steps := make([]models.TaskStep, 0)
	stepIndex := 0
	for _, pipelineStep := range task.Flow.Steps {
		taskStep := newScanTaskStep(tx, task, pipelineStep, stepIndex)
		steps = append(steps, *taskStep)
		stepIndex += 1
	}

	if len(steps) == 0 {
		return nil, e.New(e.TaskNotHaveStep, fmt.Errorf("task have no steps"))
	}

	if err := tx.Insert(&task); err != nil {
		return nil, e.New(e.DBError, errors.Wrapf(err, "save task"))
	}

	for i := range steps {
		if i+1 < len(steps) {
			steps[i].NextStep = steps[i+1].Id
		}
		if err := tx.Insert(&steps[i]); err != nil {
			return nil, e.New(e.DBError, errors.Wrapf(err, "save task step"))
		}
	}
	return &task, nil
}

func GetScanTaskById(tx *db.Session, id models.Id) (*models.ScanTask, e.Error) {
	o := models.ScanTask{}
	if err := tx.Where("id = ?", id).First(&o); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TaskNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &o, nil
}

// CreateMirrorScanTask 创建镜像扫描任务
func CreateMirrorScanTask(task *models.Task) *models.ScanTask {
	return &models.ScanTask{
		BaseTask:     task.BaseTask,
		OrgId:        task.OrgId,
		ProjectId:    task.ProjectId,
		TplId:        task.TplId,
		EnvId:        task.EnvId,
		Name:         task.Name,
		CreatorId:    task.CreatorId,
		RepoAddr:     task.RepoAddr,
		Revision:     task.Revision,
		CommitId:     task.CommitId,
		Workdir:      task.Workdir,
		Mirror:       true,
		MirrorTaskId: task.Id,
	}
}

// 查询任务所有的步骤信息
func QueryTaskStepsById(query *db.Session, taskId models.Id) *db.Session {
	return query.Model(&models.TaskStep{}).Where("task_id = ?", taskId).Order("`index`")
}

// 查询任务下某一个单独步骤的具体执行日志
func GetTaskStepLogById(tx *db.Session, stepId models.Id) ([]byte, e.Error) {
	query := tx.Joins("left join iac_task_step on iac_task_step.log_path=iac_storage.path").
		Where("iac_task_step.id = ?", stepId).
		LazySelectAppend("iac_storage.content")

	var dbStorage models.DBStorage
	if err := query.Find(&dbStorage); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return dbStorage.Content, nil
}

func SendKafkaMessage(session *db.Session, task *models.Task, taskStatus string) {
	resources := make([]models.Resource, 0)
	if err := session.Model(models.Resource{}).Where("org_id = ? AND project_id = ? AND env_id = ? AND task_id = ?",
		task.OrgId, task.ProjectId, task.EnvId, task.Id).Find(&resources); err != nil {
		logs.Get().Errorf("kafka send error, get resource data err: %v", err)
		return
	}
	k := kafka.Get()
	message := k.GenerateKafkaContent(task, taskStatus, resources)
	if err := k.ConnAndSend(message); err != nil {
		logs.Get().Errorf("kafka send error: %v", err)
		return
	}
	logs.Get().Infof("kafka send massage successful. data: %s", string(message))
}

type Resource struct {
	models.Resource
	DriftDetail string       `json:"driftDetail"`
	DriftAt     *models.Time `json:"driftAt"`
	IsDrift     bool         `json:"isDrift" form:"isDrift" `
}

func GetTaskResourceToTaskId(dbSess *db.Session, task *models.Task) ([]Resource, e.Error) {
	// 查询出最后一次漂移检测的资源
	// 资源类型: 新增、删除、修改
	rs := make([]Resource, 0)
	if err := dbSess.Table("iac_resource as r").
		Joins("left join iac_resource_drift as rd on rd.res_id =  r.id ").
		Where("r.org_id = ? AND r.project_id = ? AND r.env_id = ? AND r.task_id = ?",
			task.OrgId, task.ProjectId, task.EnvId, task.Id).
		LazySelectAppend("r.*, rd.drift_detail, rd.updated_at, rd.created_at").
		Find(&rs); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return rs, nil
}

func InsertOrUpdateCronTaskInfo(session *db.Session, resDrift models.ResourceDrift) {
	//_, err := session.Exec("replace into iac_resource_drift(id, res_id, drift_detail, created_at, update_at) values (?,?,?,?,?)",
	//	resDrift.Id, resDrift.ResId, resDrift.DriftDetail, resDrift.CreatedAt, resDrift.UpdatedAt)
	//if err != nil {
	//	logs.Get().Errorf("insert cron task info error: %v", err)
	//}
	exist, err := session.Where("res_id = ?", resDrift.ResId).Model(&resDrift).Exists()
	if err != nil {
		logs.Get().Errorf("insert resource drift info error: %v", err)
		return
	}
	if !exist {
		if err = models.Create(session, &resDrift); err != nil {
			logs.Get().Errorf("insert resource drift info error: %v", err)
		}
	} else {
		_, err = models.UpdateModelAll(session, &resDrift)
		if err != nil {
			logs.Get().Errorf("update resource drift info error: %v", err)
		}
	}
}

func SendVcsComment(session *db.Session, task *models.Task, taskStatus string) {
	env, err := GetEnvById(session, task.EnvId)
	if err != nil {
		logs.Get().Errorf("vcs comment err, get env detail data err: %v", err)
		return
	}

	vp, err := GetVcsPrByTaskId(session, task)
	if err != nil {
		logs.Get().Errorf("vcs comment err, get vcs pr data err: %v", err)
		return
	}

	vcs, err := GetVcsRepoByTplId(session, task.TplId)
	if err != nil {
		logs.Get().Errorf("vcs comment err, get vcs data err: %v", err)
		return
	}
	taskStep, err := GetTaskPlanStep(session, task.Id)
	if err != nil {
		logs.Get().Errorf("vcs comment err, get task step data err: %v", err)
		return
	}

	logContent, er := logstorage.Get().Read(taskStep.LogPath)
	if er != nil {
		logs.Get().Errorf("vcs comment err, get task plan log err: %v", er)
		return
	}

	attr := map[string]interface{}{
		"Status": taskStatus,
		"Name":   env.Name,
		//http://{{addr}}/org/{{orgId}}/project/{{ProjectId}}/m-project-env/detail/{{envId}}/task/{{TaskId}}
		"Addr":    fmt.Sprintf("%s/org/%s/project/%s/m-project-env/detail/%s/task/%s", configs.Get().Portal.Address, task.OrgId, task.ProjectId, task.EnvId, task.Id),
		"Content": stripansi.Strip(string(logContent)),
	}

	content := utils.SprintTemplate(consts.PrCommentTpl, attr)

	if err := vcs.CreatePrComment(vp.PrId, content); err != nil {
		logs.Get().Errorf("vcs comment err, create comment err: %v", err)
		return
	}
}

func QueryResource(dbSess *db.Session, task *models.Task) *db.Session {
	return dbSess.Table("iac_resource as r").
		Joins("inner join iac_resource_drift as rd on rd.address =  r.address  and rd.env_id = ? ", task.EnvId).
		Where("r.org_id = ? AND r.project_id = ? AND r.env_id = ? AND r.task_id = ?",
			task.OrgId, task.ProjectId, task.EnvId, task.Id)
}

func GetDriftResource(session *db.Session, envId, driftTaskId models.Id) ([]models.ResourceDrift, e.Error) {
	driftResources := make([]models.ResourceDrift, 0)
	if err := session.Model(&models.ResourceDrift{}).
		Where("env_id = ?", envId).
		Where("task_id = ?", driftTaskId).
		Find(&driftResources); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return driftResources, nil
}

type ResourceDriftResp struct {
	models.ResourceDrift
	IsDrift bool `gorm:"-" json:"isDrift"`
}

func GetDriftResourceById(session *db.Session, id string) (*ResourceDriftResp, e.Error) {
	driftResources := &ResourceDriftResp{}
	if err := session.Model(&models.ResourceDrift{}).
		Where("id = ?", id).
		First(driftResources); err != nil {
		return nil, e.New(e.DBError, err)
	}
	driftResources.IsDrift = true
	return driftResources, nil
}
