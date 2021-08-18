// Copyright 2021 CloudJ Company Limited. All rights reserved.

package task_manager

import (
	"cloudiac/common"
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/portal/services/logstorage"
	"cloudiac/runner"
	"cloudiac/utils"
	"cloudiac/utils/consul"
	"cloudiac/utils/logs"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

const (
	TaskManagerLockKey = "task-manager-lock"
)

var (
	ErrMaxTasksPerRunner = fmt.Errorf("concurrent limite")
)

type TaskManager struct {
	id     string
	db     *db.Session
	logger logs.Logger

	envRunningTask sync.Map       // 每个环境下正在执行的任务
	runnerTaskNum  map[string]int // 每个 runner 正在执行的任务数量

	wg sync.WaitGroup // 等待执行任务协程退出的 wait group

	maxTasksPerRunner int // 每个 runner 并发任务数量限制
}

func Start(serviceId string) {
	m := TaskManager{
		id:     serviceId,
		logger: logs.Get().WithField("worker", "taskManager").WithField("serviceId", serviceId),
	}

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					m.logger.Errorf("panic: %v", r)
					m.logger.Debugf("stack: %s", debug.Stack())
				}
			}()
			m.start()
		}()

		time.Sleep(time.Second * 10)
	}
}

func (m *TaskManager) reset() {
	m.db = db.Get()
	m.envRunningTask = sync.Map{}
	m.runnerTaskNum = make(map[string]int)
	m.wg = sync.WaitGroup{}
	m.maxTasksPerRunner = services.GetRunnerMax()
}

func (m *TaskManager) acquireLock(ctx context.Context) (<-chan struct{}, error) {
	locker, err := consul.GetLocker(TaskManagerLockKey, []byte(m.id), configs.Get().Consul.Address)
	if err != nil {
		return nil, errors.Wrap(err, "get locker")
	}

	stopLockCh := make(chan struct{})
	lockHeld := false
	go func() {
		<-ctx.Done()
		close(stopLockCh)
		if lockHeld {
			if err := locker.Unlock(); err != nil {
				m.logger.Errorf("release lock error: %v", err)
			}
		}
	}()

	lockLostCh, err := locker.Lock(stopLockCh)
	if err != nil {
		return nil, errors.Wrap(err, "acquire lock")
	}
	lockHeld = true
	return lockLostCh, nil
}

func (m *TaskManager) start() {
	m.reset()

	// ctx 用于:
	// 	1. 通知释放分布式锁
	//	2. 通知所有 task 协程退出
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		m.stop()
	}()

	m.logger.Infof("acquire task manager lock ...")
	lockLostCh, err := m.acquireLock(ctx)
	if err != nil {
		// 正常情况下 acquireLock 会阻塞直到成功获取锁，如果报错了就是出现了异常(可能是连接问题)
		m.logger.Errorf("acquire task manager lock failed: %v", err)
		return
	}

	m.logger.Infof("task manager started")

	go func() {
		<-lockLostCh
		m.logger.Infof("task manager lock lost")
		cancel()
	}()

	// 恢复执行中的任务状态
	if err = m.recoverTask(ctx); err != nil {
		m.logger.Errorf("recover task error: %v", err)
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		if err := m.processAutoDestroy(); err != nil {
			m.logger.Errorf("process auto destroy error: %v", err)
		}

		m.processPendingTask(ctx)

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			m.logger.Infof("context done: %v", ctx.Err())
			return
		}
	}
}

func (m *TaskManager) recoverTask(ctx context.Context) error {
	logger := m.logger
	query := m.db.Model(&models.Task{}).
		Where("status IN (?)", []string{models.TaskRunning, models.TaskApproving})

	tasks := make([]*models.Task, 0)
	if err := query.Find(&tasks); err != nil {
		logger.Errorf("find '%s' tasks error: %v", models.TaskRunning, err)
		return err
	}

	logger.Infof("find '%d' running tasks", len(tasks))
	for _, task := range tasks {
		select {
		case <-ctx.Done():
			break
		default:
			logger.Infof("recover running task %s", task.Id)
			if err := m.runTask(ctx, task); err != nil {
				logger.WithField("taskId", task.Id).Errorf("run task error: %s", err)
				return err
			}
		}
	}

	return nil
}

func (m *TaskManager) processPendingTask(ctx context.Context) {
	logger := m.logger

	limitedRunners := make([]string, 0)
	for runnerId, count := range m.runnerTaskNum {
		if count >= m.maxTasksPerRunner {
			limitedRunners = append(limitedRunners, runnerId)
		}
	}

	runningEnvs := make([]models.Id, 0)
	m.envRunningTask.Range(func(key, value interface{}) bool {
		runningEnvs = append(runningEnvs, key.(models.Id))
		return true
	})

	// 查询每个环境中最先创建的且状态为 pending 的任务
	firstPendingQuery := m.db.Raw(
		"SELECT env_id, MIN(created_at) AS created_at FROM iac_task "+
			"WHERE status = ? GROUP BY env_id", models.TaskPending)
	// 根据上一步查询到的 env_id + created_at 查询符合条件的任务，并取每个 env 下 id 最小的一条 task 记录
	// (注意：直接通过 env_id + created_at 匹配可能同一个 env 会返回多条记录)
	firstPendingIdQuery := m.db.Raw("SELECT iac_task.env_id, MIN(iac_task.id) AS task_id FROM iac_task, (?) AS fpt "+
		"WHERE iac_task.env_id = fpt.env_id AND iac_task.created_at = fpt.created_at "+
		"AND iac_task.status = ? GROUP BY env_id", firstPendingQuery.Expr(), models.TaskPending)

	// 通过 id 查询完整任务信息
	query := m.db.Model(&models.Task{}).Joins("JOIN (?) AS t ON t.task_id = iac_task.id", firstPendingIdQuery.Expr())

	if len(runningEnvs) > 0 {
		// 过滤掉同一环境下有其他任务在执行的任务
		query = query.Where("iac_task.env_id NOT IN (?)", runningEnvs)
	}

	if len(limitedRunners) > 0 {
		// 查询时过滤掉己达并发限制的 runner
		query = query.Where("runner_id NOT IN (?)", limitedRunners)
	}

	queryTaskLimit := 64 // 单次查询任务数量限制
	tasks := make([]*models.Task, 0)
	if err := query.Limit(queryTaskLimit).Find(&tasks); err != nil {
		logger.Panicf("find '%s' task error: %v", models.TaskPending, err)
	}

	for i := range tasks {
		select {
		case <-ctx.Done():
			break
		default:
		}

		task := tasks[i]
		// 判断 runner 并发数量
		n := m.runnerTaskNum[task.RunnerId]
		if n >= m.maxTasksPerRunner {
			logger.WithField("count", n).Infof("runner %s: %v", task.RunnerId, ErrMaxTasksPerRunner)
			continue
		}

		if err := m.runTask(ctx, task); err != nil {
			if err == errHasRunningTask {
				continue
			} else {
				logger.WithField("taskId", task.Id).Errorf("run task error: %s", err)
			}
		}
	}
}

var (
	errHasRunningTask = errors.New("environment has running task")
)

func (m *TaskManager) runTask(ctx context.Context, task *models.Task) error {
	logger := m.logger.WithField("taskId", task.Id)

	if v, loaded := m.envRunningTask.LoadOrStore(task.EnvId, task); loaded {
		t := v.(*models.Task)
		logger.Infof("environment '%s' has running task '%s'", task.EnvId, t.Id)
		return errHasRunningTask
	}

	m.wg.Add(1)
	go func() {
		defer func() {
			m.envRunningTask.Delete(task.EnvId)
			m.wg.Done()
		}()

		if startErr := m.doRunTask(ctx, task); startErr == nil {
			// 任务启动成功，执行任务结束后的处理函数
			m.processTaskDone(task)
		}
	}()
	return nil
}

// doRunTask, startErr 只在任务启动出错时(执行步骤前出错)才会返回错误
func (m *TaskManager) doRunTask(ctx context.Context, task *models.Task) (startErr error) {
	logger := m.logger.WithField("taskId", task.Id)

	changeTaskStatus := func(status, message string) error {
		if er := services.ChangeTaskStatus(m.db, task, status, message); er != nil {
			logger.Errorf("update task status error: %v", er)
			return er
		}
		return nil
	}

	taskStartFailed := func(err error) {
		logger.Infof("task failed: %s", err)
		startErr = err
		_ = changeTaskStatus(models.TaskFailed, err.Error())
	}

	logger.Infof("run task: %s", task.Id)

	if !task.Started() { // 任务可能为己启动状态(比如异常退出后的任务恢复)，这里判断一下
		// 先更新任务为 running 状态
		// 极端情况下任务未执行好过重复执行，所以先设置状态，后发起调用
		if err := changeTaskStatus(models.TaskRunning, ""); err != nil {
			return
		}
	}

	if task.IsEffectTask() {
		if _, er := m.db.Model(&models.Env{}).Where("id = ?", task.EnvId).
			Update(&models.Env{LastTaskId: task.Id}); er != nil {
			logger.Errorf("update env lastTaskId: %v", er)
			return
		}
	}

	runTaskReq, err := buildRunTaskReq(m.db, *task)
	if err != nil {
		taskStartFailed(err)
		return
	}

	steps, err := services.GetTaskSteps(m.db, task.Id)
	if err != nil {
		taskStartFailed(errors.Wrap(err, "get task steps"))
		return
	}

	var step *models.TaskStep
	for _, step = range steps {
		if step.Index < task.CurrStep {
			// 跳过己执行的步骤
			continue
		}

		if _, err = m.db.Model(task).UpdateAttrs(models.Attrs{"CurrStep": step.Index}); err != nil {
			logger.Errorf("update task error: %v", err)
			break
		}

		if err = m.runTaskStep(ctx, *runTaskReq, task, step); err != nil {
			logger.Infof("run task step: %v", err)
			break
		}
	}

	if task.IsEffectTask() && step != nil && !step.IsRejected() {
		// 执行信息采集步骤
		if err := m.runTaskStep(ctx, *runTaskReq, task, &models.TaskStep{
			TaskStepBody: models.TaskStepBody{
				Type: models.TaskStepCollect,
			},
			OrgId:     task.OrgId,
			ProjectId: task.ProjectId,
			EnvId:     task.EnvId,
			TaskId:    task.Id,
			Index:     common.CollectTaskStepIndex,
			Status:    models.TaskStepPending,
		}); err != nil {
			logger.Errorf("run collect step error: %v", err)
		} else {
			logger.Infof("collect step done")
		}
	}

	logger.Infof("task done, status: %s", task.Status)
	return nil
}

func (m *TaskManager) processTaskDone(task *models.Task) {
	logger := m.logger.WithField("func", "processTaskDone").WithField("taskId", task.Id)

	dbSess := m.db
	read := func(path string) ([]byte, error) {
		content, err := logstorage.Get().Read(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, err
		}
		return content, nil
	}

	// 分析环境资源、outputs
	processState := func() error {
		if bs, err := read(task.StateJsonPath()); err != nil {
			return fmt.Errorf("read state json: %v", err)
		} else if len(bs) > 0 {
			tfState, err := services.UnmarshalStateJson(bs)
			if err != nil {
				return fmt.Errorf("unmarshal state json: %v", err)
			}
			if err = services.SaveTaskResources(dbSess, task, tfState.Values); err != nil {
				return fmt.Errorf("save task resources: %v", err)
			}
			if err = services.SaveTaskOutputs(dbSess, task, tfState.Values.Outputs); err != nil {
				return fmt.Errorf("save task outputs: %v", err)
			}
			if err = services.UpdateEnvModel(dbSess, task.EnvId, models.Env{LastResTaskId: task.Id}); err != nil {
				return fmt.Errorf("update env lastResTaskId: %v", err)
			}
		}

		return nil
	}

	processPlan := func() error {
		if bs, err := read(task.PlanJsonPath()); err != nil {
			return fmt.Errorf("read plan json: %v", err)
		} else if len(bs) > 0 {
			tfPlan, err := services.UnmarshalPlanJson(bs)
			if err != nil {
				return fmt.Errorf("unmarshal plan json: %v", err)
			}
			if err = services.SaveTaskChanges(dbSess, task, tfPlan.ResourceChanges); err != nil {
				return fmt.Errorf("save task changes: %v", err)
			}
		}
		return nil
	}

	processTfParse := func() error {
		if bs, err := read(task.TfParseJsonPath()); err != nil {
			return fmt.Errorf("read parse json: %v", err)
		} else if len(bs) > 0 {
			_, err := services.UnmarshalTfParseJson(bs)
			if err != nil {
				return fmt.Errorf("unmarshal parse json: %v", err)
			}
			//TODO: return parse result
		}
		return nil
	}

	processTfResult := func() error {
		if bs, err := read(task.PlanJsonPath()); err != nil {
			return fmt.Errorf("read plan json: %v", err)
		} else if len(bs) > 0 {
			tfResultJson, err := services.UnmarshalTfResultJson(bs)
			if err != nil {
				return fmt.Errorf("unmarshal result json: %v", err)
			}
			if err = services.SaveTfScanResult(dbSess, task, tfResultJson.Results); err != nil {
				return fmt.Errorf("save scan result: %v", err)
			}
		}
		return nil
	}

	// 设置 auto destroy
	processAutoDestroy := func() error {
		env, err := services.GetEnv(dbSess, task.EnvId)
		if err != nil {
			return errors.Wrapf(err, "get env '%s'", task.EnvId)
		}

		updateAttrs := models.Attrs{}

		if task.Type == models.TaskTypeDestroy && env.Status == models.EnvStatusInactive {
			// 环境销毁后清空自动销毁设置，以支持通过再次部署重建环境。
			// ttl 需要保留，做为重建环境的默认 ttl
			updateAttrs["AutoDestroyAt"] = nil
			updateAttrs["AutoDestroyTaskId"] = ""
		}

		// 如果设置了环境的 ttl，则在部署成功后自动根据 ttl 设置销毁时间。
		// 该逻辑只在环境从非活跃状态变为活跃时执行，活跃环境修改 ttl 会立即计算 AutoDestroyAt
		if task.Type == models.TaskTypeApply && env.Status == models.EnvStatusActive &&
			env.AutoDestroyAt == nil && env.TTL != "" && env.TTL != "0" {
			ttl, err := services.ParseTTL(env.TTL)
			if err != nil {
				return err
			}
			at := models.Time(time.Now().Add(ttl))
			updateAttrs["AutoDestroyAt"] = &at
		}

		_, err = services.UpdateEnv(dbSess, env.Id, updateAttrs)
		if err != nil {
			return errors.Wrapf(err, "update environment")
		}

		return nil
	}

	lastStep, err := services.GetTaskStep(dbSess, task.Id, task.CurrStep)
	if err != nil {
		logger.Errorf("get task step(%d) error: %v", err, task.CurrStep)
		return
	}

	// 基于最后一个步骤更新任务状态
	updateTaskStatus := func() error {
		if err = services.ChangeTaskStatusWithStep(dbSess, task, lastStep); err != nil {
			return err
		}
		return nil
	}

	if lastStep.Type == models.TaskStepTfParse {
		if err := processTfParse(); err != nil {
			logger.Errorf("process task parse: %s", err)
		}
	} else if lastStep.Type == models.TaskStepTfScan {
		if err := processTfResult(); err != nil {
			logger.Errorf("process task scan: %s", err)
		}
	}
	if !lastStep.IsRejected() { // 任务被审批驳回时会即时更新状态，且不会执行资源统计步骤
		if task.IsEffectTask() {
			if err := processState(); err != nil {
				logger.Errorf("process task state: %v", err)
			}

			// 任务执行成功才会进行 changes 统计，失败的话基于 plan 文件进行变更统计是不准确的
			// (terraform 执行 apply 失败也不会输出资源变更情况)
			if lastStep.Status == models.TaskComplete {
				if err := processPlan(); err != nil {
					logger.Errorf("process task plan: %v", err)
				}
			}
		}

		if err := updateTaskStatus(); err != nil {
			logger.Errorf("update task status error: %v", err)
		}

		if task.IsEffectTask() {
			// 注意: 该步骤需要在环境状态被更新之后执行
			if err := processAutoDestroy(); err != nil {
				logger.Errorf("process auto destroy: %v", err)
			}
		}
	}
}

func (m *TaskManager) runTaskStep(ctx context.Context, taskReq runner.RunTaskReq,
	task *models.Task, step *models.TaskStep) (err error) {
	logger := m.logger.WithField("taskId", taskReq.TaskId)
	logger = logger.WithField("func", "runTaskStep").
		WithField("step", fmt.Sprintf("%d(%s)", step.Index, step.Type))

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("run task step painc: %v", r)
			logger.Errorln(err)
			logger.Debugf("%s", debug.Stack())
		}
	}()

	changeStepStatus := func(status, message string) {
		var er error
		if er = services.ChangeTaskStepStatus(m.db, task, step, status, message); er != nil {
			er = errors.Wrap(er, "update step status error")
			logger.Error(er)
			panic(err)
		}
	}

	if !task.AutoApprove && !step.IsApproved() {
		logger.Infof("waitting task step approve")
		changeStepStatus(models.TaskStepApproving, "")

		var newStep *models.TaskStep
		if newStep, err = WaitTaskStepApprove(ctx, m.db, step.TaskId, step.Index); err != nil {
			if err == context.Canceled {
				return err
			}

			logger.Errorf("wait task step approve error: %v", err)
			status := models.TaskStepFailed
			if err == ErrTaskStepRejected {
				status = models.TaskStepRejected
			}
			changeStepStatus(status, err.Error())
			return err
		}
		step = newStep
	}

loop:
	for {
		select {
		case <-ctx.Done():
			logger.Infof("context done")
			return ctx.Err()
		default:
		}

		switch step.Status {
		case models.TaskStepPending, models.TaskStepApproving:
			// 先将步骤置为 running 状态，然后再发起调用，保证步骤不会重复执行
			changeStepStatus(models.TaskStepRunning, "")
			logger.Infof("start task step %d(%s)", step.Index, step.Type)
			if err = StartTaskStep(taskReq, *step); err != nil {
				logger.Errorf("start task step error: %s", err.Error())
				changeStepStatus(models.TaskStepFailed, err.Error())
				return err
			}
		case models.TaskStepRunning:
			if _, err = WaitTaskStep(ctx, m.db, task, step); err != nil {
				logger.Errorf("wait task result error: %v", err)
				changeStepStatus(models.TaskStepFailed, err.Error())
				return err
			}
		default:
			break loop
		}
	}

	switch step.Status {
	case models.TaskStepComplete:
		return nil
	case models.TaskStepFailed:
		if step.Message != "" {
			return fmt.Errorf(step.Message)
		}
		return errors.New("failed")
	case models.TaskStepTimeout:
		return errors.New("timeout")
	default:
		return fmt.Errorf("unknown step status: %v", step.Status)
	}
}

func (m *TaskManager) stop() {
	logger := m.logger
	logger.Infof("task manager stopping ...")

	logger.Debugf("waiting all task goroutine exit ...")
	m.wg.Wait()
	logger.Infof("task manager stopped")
}

// buildRunTaskReq 基于任务信息构建一个 RunTaskReq 对象。
// 	注意这里不会设置 step 相关的数据，step 相关字段在 StartTaskStep() 方法中设置
func buildRunTaskReq(dbSess *db.Session, task models.Task) (taskReq *runner.RunTaskReq, err error) {
	runnerEnv := runner.TaskEnv{
		Id:              string(task.EnvId),
		Workdir:         task.Workdir,
		TfVarsFile:      task.TfVarsFile,
		Playbook:        task.Playbook,
		PlayVarsFile:    task.PlayVarsFile,
		EnvironmentVars: make(map[string]string),
		TerraformVars:   make(map[string]string),
		AnsibleVars:     make(map[string]string),
	}

	for _, v := range task.Variables {
		value := utils.EncodeSecretVar(v.Value, v.Sensitive)
		switch v.Type {
		case consts.VarTypeEnv:
			runnerEnv.EnvironmentVars[v.Name] = value
		case consts.VarTypeTerraform:
			runnerEnv.TerraformVars[v.Name] = value
		case consts.VarTypeAnsible:
			runnerEnv.AnsibleVars[v.Name] = value
		default:
			return nil, fmt.Errorf("unknown variable type: %s", v.Type)
		}
	}

	stateStore := runner.StateStore{
		Backend: "consul",
		Scheme:  "http",
		Path:    task.StatePath,
		Address: "",
	}

	pk := ""
	if task.KeyId != "" {
		mKey, err := services.GetKeyById(dbSess, task.KeyId, false)
		if err != nil {
			return nil, errors.Wrapf(err, "get key '%s' error: %v", task.KeyId, err)
		}
		pk = mKey.Content
	}

	policies, err := services.GetTaskPolicies(dbSess, task.Id)
	if err != nil {
		return nil, errors.Wrapf(err, "get task '%s' policies error: %v", task.Id, err)
	}

	taskReq = &runner.RunTaskReq{
		Env:             runnerEnv,
		RunnerId:        task.RunnerId,
		TaskId:          string(task.Id),
		DockerImage:     "",
		StateStore:      stateStore,
		RepoAddress:     task.RepoAddr,
		RepoRevision:    task.CommitId,
		Timeout:         task.StepTimeout,
		Policies:        policies,
		StopOnViolation: task.StopOnViolation,
	}
	if pk != "" {
		taskReq.PrivateKey = utils.EncodeSecretVar(pk, true)
	}

	return taskReq, nil
}

func (m *TaskManager) processAutoDestroy() error {
	logger := m.logger.WithField("func", "processAutoDestroy")

	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("panic: %v", r)
			logger.Debugf("%s", debug.Stack())
		}
	}()

	dbSess := m.db
	limit := 64
	destroyEnvs := make([]*models.Env, 0, limit)
	err := dbSess.Model(&models.Env{}).
		Where("status IN (?)", []string{models.EnvStatusActive, models.EnvStatusFailed}).
		Where("auto_destroy_task_id = ''").
		Where("auto_destroy_at <= ?", time.Now()).
		Order("auto_destroy_at").Limit(limit).Find(&destroyEnvs)

	if err != nil {
		return errors.Wrapf(err, "query destroy task: %v", err)
	}

	for _, env := range destroyEnvs {
		err = func() error {
			logger := logger.WithField("envId", env.Id)

			tx := dbSess.Begin()
			defer func() {
				if r := recover(); r != nil {
					_ = tx.Rollback()
					panic(r)
				}
			}()

			tpl, err := services.GetTemplateById(tx, env.TplId)
			if err != nil {
				_ = tx.Rollback()
				if e.IsRecordNotFound(err) {
					logger.Warnf("template %s not exists", tpl.Id)
					return nil
				} else {
					logger.Errorf("get template %s error: %v", tpl.Id, err)
					return err
				}
			}

			vars, err, _ := services.GetValidVariables(tx, consts.ScopeEnv, env.OrgId, env.ProjectId, env.TplId, env.Id, true)
			if err != nil {
				logger.Errorf("get vairables error: %v", err)
				return nil
			}

			taskVars := services.GetVariableBody(vars)

			task, err := services.CreateTask(tx, tpl, env, models.Task{
				Name:            "Auto Destroy",
				Type:            models.TaskTypeDestroy,
				Flow:            models.TaskFlow{},
				Targets:         nil,
				CreatorId:       consts.SysUserId,
				RunnerId:        "",
				Variables:       taskVars,
				StepTimeout:     0,
				AutoApprove:     true,
				StopOnViolation: env.StopOnViolation,
			})
			if err != nil {
				_ = tx.Rollback()
				logger.Errorf("create task error: %v", err)
				// 创建任务失败继续处理其他任务
				return nil
			}

			if _, err := tx.Model(&models.Env{}).Where("id = ?", env.Id).
				Update(&models.Env{AutoDestroyTaskId: task.Id}); err != nil {
				_ = tx.Rollback()
				logger.Errorf("update env error: %v", err)
				return nil
			}

			if err := tx.Commit(); err != nil {
				logger.Errorf("commit error: %v", err)
				return err
			}

			logger.Infof("created auto destory task: %s", task.Id)
			return nil
		}()

		if err != nil {
			break
		}
	}

	return nil
}
