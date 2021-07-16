package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services/logstorage"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"io"
	"net/url"
	"os"
	"strings"
	"time"
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

func CreateTask(tx *db.Session, tpl *models.Template, env *models.Env, pt models.Task) (*models.Task, e.Error) {
	logger := logs.Get().WithField("func", "CreateTask")

	var (
		err      error
		firstVal = utils.FirstValueStr
	)

	task := models.Task{
		// 以下为需要外部传入的属性
		Name:        pt.Name,
		Type:        pt.Type,
		Flow:        pt.Flow,
		Targets:     pt.Targets,
		CommitId:    pt.CommitId,
		CreatorId:   pt.CreatorId,
		RunnerId:    firstVal(pt.RunnerId, env.RunnerId),
		Variables:   pt.Variables,
		StepTimeout: pt.StepTimeout,
		AutoApprove: pt.AutoApprove,
		Extra:       pt.Extra,

		OrgId:     env.OrgId,
		ProjectId: env.ProjectId,
		TplId:     env.TplId,
		EnvId:     env.Id,
		StatePath: env.StatePath,

		RepoAddr:     "",
		Workdir:      firstVal(tpl.Workdir),
		Playbook:     firstVal(env.Playbook, tpl.Playbook),
		TfVarsFile:   firstVal(env.TfVarsFile, tpl.TfVarsFile),
		PlayVarsFile: firstVal(env.PlayVarsFile, tpl.PlayVarsFile),

		Status:   models.TaskPending,
		Message:  "",
		CurrStep: 0,
	}

	task.Id = models.NewId("run")
	logger = logger.WithField("taskId", task.Id)

	if len(task.Flow.Steps) == 0 {
		task.Flow, err = models.DefaultTaskFlow(task.Type)
		if err != nil {
			return nil, e.New(e.InternalError, err)
		}
	}

	task.RepoAddr, err = buildFullRepoAddr(tx, tpl)
	if err != nil {
		return nil, e.New(e.InternalError, err)
	}

	if _, err = tx.Save(&task); err != nil {
		return nil, e.New(e.DBError, errors.Wrapf(err, "save task"))
	}

	flowSteps := make([]models.TaskStepBody, 0, len(task.Flow.Steps))
	for i := range task.Flow.Steps {
		step := task.Flow.Steps[i]
		if step.Type == models.TaskStepPlay && task.Playbook == "" {
			logger.WithField("step", fmt.Sprintf("%d(%s)", i, step.Type)).
				Infof("not have playbook, skip this step")
			continue
		}
		flowSteps = append(flowSteps, step)
	}

	if len(flowSteps) == 0 {
		return nil, e.New(e.TaskNotHaveStep, fmt.Errorf("task have no steps"))
	}

	var preStep *models.TaskStep
	for i := len(flowSteps) - 1; i >= 0; i-- { // 倒序保存 steps，以便于设置 step.NextStep
		step := flowSteps[i]

		if len(task.Targets) != 0 && IsTerraformStep(step.Type) {
			if step.Type != models.TaskStepInit {
				for _, t := range task.Targets {
					step.Args = append(step.Args, fmt.Sprintf("-target=%s", t))
				}
			}
			// TODO: tfVars, playVars 也以这种方式传入？
		}

		nextStepId := models.Id("")
		if preStep != nil {
			nextStepId = preStep.Id
		}
		var er e.Error
		preStep, er = createTaskStep(tx, task, step, i, nextStepId)
		if er != nil {
			return nil, e.New(er.Code(), errors.Wrapf(er, "save task step"))
		}
	}

	return &task, nil
}

func buildFullRepoAddr(tx *db.Session, tpl *models.Template) (string, error) {
	var (
		u         *url.URL
		err       error
		repoAddr  = tpl.RepoAddr
		repoToken = tpl.RepoToken
	)

	if (repoToken == "" || !strings.Contains("://", repoAddr)) && tpl.VcsId != "" {
		var vcs *models.Vcs
		if vcs, err = QueryVcsByVcsId(tpl.VcsId, tx); err != nil {
			if e.IsRecordNotFound(err) {
				return "", e.New(e.VcsNotExists, err)
			}
			return "", e.New(e.DBError, err)
		}
		// 模板里保存的是路径，需要与 vcs.Address 组合成 URL
		if !strings.Contains("://", repoAddr) {
			repoAddr = utils.JoinURL(vcs.Address, tpl.RepoAddr)
		}
		// 模板没有保存 token，使用 vcs 的 token
		if repoToken == "" {
			repoToken = vcs.VcsToken
		}
	}

	u, err = url.Parse(repoAddr)
	if err != nil {
		return "", e.New(e.InternalError, errors.Wrapf(err, "parse url: %v", repoAddr))
	} else if repoToken != "" {
		u.User = url.UserPassword("token", repoToken)
	}
	return u.String(), nil
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
	return query.Model(&models.Task{})
}

var stepStatus2TaskStatus = map[string]string{
	models.TaskStepPending:   models.TaskPending,
	models.TaskStepApproving: models.TaskApproving,
	models.TaskStepRejected:  models.TaskFailed,
	models.TaskStepRunning:   models.TaskRunning,
	models.TaskStepFailed:    models.TaskFailed,
	models.TaskStepTimeout:   models.TaskFailed,
	// complete 状态需要特殊处理
}

func ChangeTaskStatusWithStep(dbSess *db.Session, task *models.Task, step *models.TaskStep) e.Error {
	var (
		taskStatus string
	)

	if step.Status == models.TaskStepComplete {
		if step.NextStep == "" {
			taskStatus = models.TaskComplete
		} else {
			taskStatus = models.TaskRunning
		}
	} else {
		var ok bool
		taskStatus, ok = stepStatus2TaskStatus[step.Status]
		if !ok {
			panic(fmt.Errorf("unknown task step status %v", step.Status))
		}
	}

	return ChangeTaskStatus(dbSess, task, taskStatus, step.Message)
}

// ChangeTaskStatus 修改任务状态(同步修改 StartAt、EndAt 等)，并同步修改 env 状态
func ChangeTaskStatus(dbSess *db.Session, task *models.Task, status, message string) e.Error {
	if task.Status == status && message == "" {
		return nil
	}

	task.Status = status
	task.Message = message
	now := time.Now()
	if task.StartAt == nil && task.Started() {
		task.StartAt = &now
	}
	if task.EndAt == nil && task.Exited() {
		task.EndAt = &now
	}

	logs.Get().WithField("taskId", task.Id).Infof("change task to '%s'", status)
	if _, err := dbSess.Model(&models.Task{}).Update(task); err != nil {
		return e.AutoNew(err, e.DBError)
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
	ProviderName string `json:"provider_name"`
	Address      string `json:"address"`
	Mode         string `json:"mode"` // managed、data
	Type         string `json:"type"`
	Name         string `json:"name"`
	Index        int    `json:"index"`

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
		rs = append(rs, &models.Resource{
			Provider: r.ProviderName,
			Module:   moduleName,
			Address:  r.Address,
			Type:     r.Type,
			Name:     r.Name,
			Index:    r.Index,
			Attrs:    r.Values,
		})
	}

	for i := range module.ChildModules {
		rs = append(rs, traverseStateModule(&module.ChildModules[i])...)
	}
	return rs
}

func SaveTaskResources(tx *db.Session, task *models.Task, values TfStateValues) error {
	bq := utils.NewBatchSQL(1024, "INSERT INTO", models.Resource{}.TableName(),
		"id", "org_id", "project_id", "env_id", "task_id",
		"provider", "module", "address", "type", "name", "index", "attrs")

	rs := make([]*models.Resource, 0)
	rs = append(rs, traverseStateModule(&values.RootModule)...)
	for i := range values.ChildModules {
		rs = append(rs, traverseStateModule(&values.ChildModules[i])...)
	}

	for _, r := range rs {
		err := bq.AddRow(models.NewId("r"), task.OrgId, task.ProjectId, task.EnvId, task.Id,
			r.Provider, r.Module, r.Address, r.Type, r.Name, r.Index, r.Attrs)
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
		Update("result", task.Result); err != nil {
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

func SaveTaskChanges(dbSess *db.Session, task *models.Task, rs []TfPlanResource) error {
	task.Result.ResAdded = 0
	task.Result.ResChanged = 0
	task.Result.ResDestroyed = 0
	for _, r := range rs {
		actions := r.Change.Actions
		switch {
		case utils.SliceEqualStr(actions, []string{"no-op"}),
			utils.SliceEqualStr(actions, []string{"create", "delete"}):
			continue
		case utils.SliceEqualStr(actions, []string{"create"}):
			task.Result.ResAdded += 1
		case utils.SliceEqualStr(actions, []string{"update"}),
			utils.SliceEqualStr(actions, []string{"delete", "create"}):
			task.Result.ResChanged += 1
		case utils.SliceEqualStr(actions, []string{"delete"}):
			task.Result.ResDestroyed += 1
		default:
			logs.Get().WithField("taskId", task.Id).Errorf("unknown change actions: %v", actions)
		}
	}

	if _, err := dbSess.Model(&models.Task{}).Where("id = ?", task.Id).
		Update("result", task.Result); err != nil {
		return err
	}
	return nil
}

func FetchTaskLog(ctx context.Context, task *models.Task, writer io.WriteCloser) (err error) {
	// close 后 read 端会触发 EOF error
	defer writer.Close()

	var steps []*models.TaskStep
	steps, err = GetTaskSteps(db.Get(), task.Id)
	if err != nil {
		return err
	}

	storage := logstorage.Get()
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for _, step := range steps {
		// 任务有可能未开始执行步骤就退出了，所以需要先判断任务是否退出
		for !task.Exited() && !step.IsStarted() {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				task, err = GetTask(db.Get(), task.Id)
				if err != nil {
					return errors.Wrapf(err, "get task '%s'", task.Id)
				}

				step, err = GetTaskStep(db.Get(), task.Id, step.Index)
				if err != nil {
					return errors.Wrapf(err, "get task step %d", step.Index)
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
			if err = fetchRunnerTaskStepLog(ctx, task.RunnerId, step, writer); err != nil {
				return err
			}
		}
	}

	return nil
}

// 从 runner 获取任务日志，直到任务结束
func fetchRunnerTaskStepLog(ctx context.Context, runnerId string, step *models.TaskStep, writer io.Writer) error {
	logger := logs.Get().WithField("func", "fetchRunnerTaskStepLog").
		WithField("taskId", step.TaskId).
		WithField("step", fmt.Sprintf("%d(%s)", step.Index, step.Type))

	runnerAddr, err := GetRunnerAddress(runnerId)
	if err != nil {
		return errors.Wrapf(err, "get runner address")
	}

	params := url.Values{}
	params.Add("envId", string(step.EnvId))
	params.Add("taskId", string(step.TaskId))
	params.Add("step", fmt.Sprintf("%d", step.Index))
	wsConn, _, err := utils.WebsocketDail(runnerAddr, consts.RunnerTaskLogFollowURL, params)
	if err != nil {
		return errors.Wrapf(err, "websocket dail: %v/%s", runnerAddr, consts.RunnerTaskLogFollowURL)
	}

	defer func() {
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
				logger.Errorf("read message error: %v", err)
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
