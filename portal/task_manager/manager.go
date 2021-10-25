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
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/pkg/errors"
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
	query := m.db.Where("status IN (?)", []string{models.TaskRunning, models.TaskApproving})

	deployTasks := make([]*models.Task, 0)
	if err := query.Model(&models.Task{}).Find(&deployTasks); err != nil {
		logger.Errorf("find '%s' deploy tasks error: %v", models.TaskRunning, err)
		return err
	}
	scanTasks := make([]*models.ScanTask, 0)
	if err := query.Model(&models.ScanTask{}).Where("mirror = 0").Find(&scanTasks); err != nil {
		logger.Errorf("find '%s' scan tasks error: %v", models.TaskRunning, err)
		return err
	}

	tasks := make([]models.Tasker, len(scanTasks)+len(deployTasks))
	// 合并等待任务列表，扫描任务更轻量，我们先执行扫描任务
	for idx := range scanTasks {
		tasks[idx] = scanTasks[idx]
	}
	scanTasksLen := len(scanTasks)
	for idx := range deployTasks {
		tasks[scanTasksLen+idx] = deployTasks[idx]
	}

	logger.Infof("find '%d' running tasks", len(tasks))
	for _, task := range tasks {
		select {
		case <-ctx.Done():
			break
		default:
			logger.Infof("recover running task %s", task.GetId())
			if err := m.runTask(ctx, task); err != nil {
				logger.WithField("taskId", task.GetId()).Errorf("run task error: %s", err)
				return err
			}
		}
	}

	return nil
}

func (m *TaskManager) getPendingDeployTasks() []*models.Task {
	logger := m.logger

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

	limitedRunners := m.getLimitedRunner()
	if len(limitedRunners) > 0 {
		// 查询时过滤掉己达并发限制的 runner
		query = query.Where("runner_id NOT IN (?)", limitedRunners)
	}

	queryTaskLimit := 64 // 单次查询任务数量限制
	tasks := make([]*models.Task, 0)
	if err := query.Limit(queryTaskLimit).Find(&tasks); err != nil {
		logger.Panicf("find '%s' task error: %v", models.TaskPending, err)
	}

	return tasks
}

func (m *TaskManager) getPendingScanTasks() []*models.ScanTask {
	logger := m.logger

	// 扫描类型任务支持多个并行执行，不会互相影响，这里获取所有处于 pending 状态的任务列表
	query := m.db.Model(&models.ScanTask{}).Where("status = ? AND mirror = 0", models.TaskPending)

	limitedRunners := m.getLimitedRunner()
	if len(limitedRunners) > 0 {
		// 查询时过滤掉己达并发限制的 runner
		query = query.Where("runner_id NOT IN (?)", limitedRunners)
	}

	queryTaskLimit := 64 // 单次查询任务数量限制
	tasks := make([]*models.ScanTask, 0)
	if err := query.Limit(queryTaskLimit).Find(&tasks); err != nil {
		logger.Panicf("find '%s' task error: %v", models.TaskPending, err)
	}

	return tasks
}

func (m *TaskManager) getLimitedRunner() []string {
	limitedRunners := make([]string, 0)
	for runnerId, count := range m.runnerTaskNum {
		if count >= m.maxTasksPerRunner {
			limitedRunners = append(limitedRunners, runnerId)
		}
	}
	return limitedRunners
}

func (m *TaskManager) processPendingTask(ctx context.Context) {
	logger := m.logger

	scanTasks := m.getPendingScanTasks()
	deployTasks := m.getPendingDeployTasks()
	tasks := make([]models.Tasker, len(scanTasks)+len(deployTasks))

	// 合并等待任务列表，扫描任务更轻量，我们先执行扫描任务
	for idx := range scanTasks {
		tasks[idx] = scanTasks[idx]
	}
	scanTasksLen := len(scanTasks)
	for idx := range deployTasks {
		tasks[scanTasksLen+idx] = deployTasks[idx]
	}

	for i := range tasks {
		select {
		case <-ctx.Done():
			break
		default:
		}

		task := tasks[i]
		// 判断 runner 并发数量
		n := m.runnerTaskNum[task.GetRunnerId()]
		if n >= m.maxTasksPerRunner {
			logger.WithField("count", n).Infof("runner %s: %v", task.GetRunnerId(), ErrMaxTasksPerRunner)
			continue
		}

		if err := m.runTask(ctx, task); err != nil {
			if err == errHasRunningTask {
				continue
			} else {
				logger.WithField("taskId", task.GetId()).Errorf("run task error: %s", err)
			}
		}
	}
}

var (
	errHasRunningTask = errors.New("environment has running task")
)

func (m *TaskManager) runTask(ctx context.Context, task models.Tasker) error {
	logger := m.logger.WithField("taskId", task.GetId())

	if t, ok := task.(*models.Task); ok {
		if v, loaded := m.envRunningTask.LoadOrStore(t.EnvId, t); loaded {
			t := v.(*models.Task)
			logger.Infof("environment '%s' has running task '%s'", t.EnvId, t.Id)
			return errHasRunningTask
		}
	}

	m.wg.Add(1)
	go func() {
		defer func() {
			if t, ok := task.(*models.Task); ok {
				m.envRunningTask.Delete(t.EnvId)
			}
			m.wg.Done()
		}()

		switch t := task.(type) {
		case *models.Task:
			if startErr := m.doRunTask(ctx, t); startErr == nil {
				// 任务启动成功，执行任务结束后的处理函数
				m.processTaskDone(t.Id)
			}
		case *models.ScanTask:
			if startErr := m.doRunScanTask(ctx, t); startErr == nil {
				m.processScanTaskDone(t.Id)
			}
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
			taskStartFailed(errors.Wrap(err, "update task"))
			return
		}

		{
			// 获取 task 最新的 containerId
			tTask, err := services.GetTaskById(m.db, task.Id)
			if err != nil {
				taskStartFailed(errors.Wrapf(err, "get task %s", task.Id.String()))
				return
			}
			runTaskReq.ContainerId = tTask.ContainerId
		}

		runErr := m.runTaskStep(ctx, *runTaskReq, task, step)
		if err := m.processStepDone(task, step); err != nil {
			logger.Infof("process step done: %v", err)
			break
		}

		if runErr != nil {
			if step.Type == common.TaskStepOpaScan && !task.StopOnViolation {
				// 合规任务失败不影响环境部署流程
				logger.Warnf("run scan task step: %v", runErr)
				continue
			}
			logger.Infof("run task step: %v", runErr)
			break
		}
	}

	if task.IsEffectTask() && step != nil && !step.IsRejected() {
		// 执行信息采集步骤
		if err := m.runTaskStep(ctx, *runTaskReq, task, &models.TaskStep{
			PipelineStep: models.PipelineStep{
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

	logger.Infof("run task finish")
	return nil
}

func (m *TaskManager) processStepDone(task *models.Task, step *models.TaskStep) error {
	dbSess := m.db
	processScanResult := func() error {
		var (
			tsResult services.TsResult
			scanStep *models.TaskStep
			scanTask *models.ScanTask
			err      error
		)

		if scanStep, err = services.GetTaskScanStep(dbSess, task.Id); err != nil {
			return err
		}
		if scanTask, err = services.GetMirrorScanTask(dbSess, task.Id); err != nil {
			return err
		}

		// 根据扫描步骤的执行结果更新扫描任务的状态
		if err = services.ChangeTaskStatusWithStep(dbSess, scanTask, scanStep); err != nil {
			return err
		}

		// 处理扫描结果
		if scanTask.PolicyStatus == common.PolicyStatusPassed || scanTask.PolicyStatus == common.PolicyStatusViolated {
			if er := services.InitScanResult(dbSess, scanTask); er != nil {
				return er
			}
			if bs, er := readIfExist(task.TfResultJsonPath()); er == nil && len(bs) > 0 {
				if tfResultJson, er := services.UnmarshalTfResultJson(bs); er == nil {
					tsResult = tfResultJson.Results
				}
			}

			if err := services.UpdateScanResult(dbSess, scanTask, tsResult, scanTask.PolicyStatus); err != nil {
				return fmt.Errorf("save scan result: %v", err)
			}
		}

		return err
	}

	switch step.Type {
	case common.TaskStepOpaScan:
		return processScanResult()
	}
	return nil
}

func readIfExist(path string) ([]byte, error) {
	content, err := logstorage.Get().Read(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return content, nil
}

func (m *TaskManager) processTaskDone(taskId models.Id) {
	logger := m.logger.WithField("func", "processTaskDone").WithField("taskId", taskId)
	logger.Debugln("start process task done")

	dbSess := m.db

	// 重新查询获取 task，确保使用的是最新的 task 数据
	task, err := services.GetTaskById(dbSess, taskId)
	if err != nil {
		logger.Errorf("get task %d: %v", err)
		return
	}

	// 分析环境资源、outputs
	processState := func() error {
		if bs, err := readIfExist(task.StateJsonPath()); err != nil {
			return fmt.Errorf("read state json: %v", err)
		} else if len(bs) > 0 {
			tfState, err := services.UnmarshalStateJson(bs)

			if err != nil {
				return fmt.Errorf("unmarshal state json: %v", err)
			}
			ps, err := readIfExist(task.ProviderSchemaJsonPath())
			proMap := runner.ProviderSensitiveAttrMap{}
			if err != nil {
				return fmt.Errorf("read provider schema json: %v", err)
			}
			if len(ps) > 0 {
				if err = json.Unmarshal(ps, &proMap); err != nil {
					return err
				}
			}
			if err = services.SaveTaskResources(dbSess, task, tfState.Values, proMap); err != nil {
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
		if bs, err := readIfExist(task.PlanJsonPath()); err != nil {
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

	if err := StopTaskContainers(dbSess, task.Id); err != nil {
		logger.Warnf("stop task containers: %v", err)
	}

	lastStep, err := services.GetTaskStep(dbSess, task.Id, task.CurrStep)
	if err != nil {
		logger.Errorf("get task step(%d) error: %v", err, task.CurrStep)
		return
	}

	logger.Debugf("last step %#v", lastStep)

	// 基于最后一个步骤更新任务状态
	updateTaskStatus := func() error {
		if err = services.ChangeTaskStatusWithStep(dbSess, task, lastStep); err != nil {
			return err
		}
		return nil
	}

	// 任务被审批驳回时会即时更新状态，且不会执行资源统计步骤，所以不需要执行下面这段逻辑
	if !lastStep.IsRejected() {
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

	if step.NextStep != "" {
		if nextStep, err := services.GetTaskStepByStepId(m.db, step.NextStep); err != nil {
			err = errors.Wrapf(err, "get task step %s", string(step.NextStep))
			return err
		} else if nextStep.MustApproval {
			taskReq.PauseTask = true
		}
	}

	changeStepStatusAndStepRetryTimes := func(status, message string, step *models.TaskStep) {
		var er error
		if er = services.ChangeTaskStepStatusAndUpdate(m.db, task, step, status, message); er != nil {
			er = errors.Wrap(er, "update step status error")
			logger.Error(er)
			panic(err)
		}
	}

	if step.MustApproval && !step.IsApproved() {
		logger.Infof("waitting task step approve")
		changeStepStatusAndStepRetryTimes(models.TaskStepApproving, "", step)
		var newStep *models.TaskStep
		if newStep, err = WaitTaskStepApprove(ctx, m.db, step.TaskId, step.Index); err != nil {
			if err == context.Canceled {
				return err
			}

			logger.Errorf("wait task step approve error: %v", err)
			if err != ErrTaskStepRejected {
				changeStepStatusAndStepRetryTimes(models.TaskStepFailed, err.Error(), step)
			}
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
		if step.NextRetryTime != 0 {
			sleepTime := step.NextRetryTime - time.Now().Unix()
			// 如果没有到达重试时间，则时间等待缺少时间
			if sleepTime > 0 {
				time.Sleep(time.Duration(sleepTime) * time.Second)
			}
		}

		switch step.Status {
		case models.TaskStepPending, models.TaskApproving:
			// 先将步骤置为 running 状态，然后再发起调用，保证步骤不会重复执行
			changeStepStatusAndStepRetryTimes(models.TaskStepRunning, "", step)
			if cid, retryAble, err := StartTaskStep(taskReq, *step); err != nil {
				logger.Warnf("start task step %d-%s: %v", step.Index, step.Type, err)
				// 如果是可重试错误，并且任务设定可以重试, 则运行重试逻辑
				if retryAble && task.RetryAble {
					if step.RetryNumber > 0 && step.CurrentRetryCount < step.RetryNumber {
						// 下次重试时间为当前任务失败时间点加任务设置重试间隔时间。
						step.NextRetryTime = time.Now().Unix() + int64(task.RetryDelay)
						step.CurrentRetryCount += 1
						message := fmt.Sprintf("Task step start failed and try again. The current number of retries is %d", step.CurrentRetryCount)
						changeStepStatusAndStepRetryTimes(models.TaskStepPending, message, step)
					}
				} else {
					changeStepStatusAndStepRetryTimes(models.TaskStepFailed, err.Error(), step)
					return err
				}
			} else if taskReq.ContainerId == "" {
				if err := services.UpdateTaskContainerId(m.db, models.Id(taskReq.TaskId), cid); err != nil {
					panic(errors.Wrapf(err, "update task %s container id", taskReq.TaskId))
				}
			}
		case models.TaskStepRunning:
			stepResult, err := WaitTaskStep(ctx, m.db, task, step)
			if err != nil {
				logger.Errorf("wait task result error: %v", err)
				changeStepStatusAndStepRetryTimes(models.TaskStepFailed, err.Error(), step)
				return err
			}
			// 合规检测步骤不通过，不需要重试，跳出循环
			if step.Type == models.TaskStepOpaScan &&
				stepResult.Result.ExitCode == common.TaskStepPolicyViolationExitCode {
				message := "Scan task step finished with violations found."
				changeStepStatusAndStepRetryTimes(models.TaskStepFailed, message, step)
				break loop
			}
			if stepResult.Status == models.TaskStepFailed {
				if task.RetryAble && step.RetryNumber > 0 && step.CurrentRetryCount < step.RetryNumber {
					step.NextRetryTime = time.Now().Unix() + int64(task.RetryDelay)
					step.CurrentRetryCount += 1
					message := fmt.Sprintf("Task step start failed and try again. The current number of retries is %d", step.CurrentRetryCount)
					changeStepStatusAndStepRetryTimes(models.TaskStepPending, message, step)
				}
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
		TfVersion:       task.TfVersion,
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

	taskReq = &runner.RunTaskReq{
		Env:             runnerEnv,
		RunnerId:        task.RunnerId,
		TaskId:          string(task.Id),
		DockerImage:     task.Flow.Image,
		StateStore:      stateStore,
		RepoAddress:     task.RepoAddr,
		RepoRevision:    task.CommitId,
		Timeout:         task.StepTimeout,
		StopOnViolation: task.StopOnViolation,
		ContainerId:     task.ContainerId,
	}
	if scanStep, err := services.GetTaskScanStep(dbSess, task.Id); err == nil && scanStep != nil {
		policies, err := services.GetTaskPolicies(dbSess, task.Id)
		if err != nil {
			return nil, errors.Wrapf(err, "get task '%s' policies error: %v", task.Id, err)
		}
		taskReq.Policies = policies
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
					logger.Warnf("template %s not exists", env.TplId)
					return nil
				} else {
					logger.Errorf("get template %s error: %v", env.TplId, err)
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
				Targets:         nil,
				CreatorId:       consts.SysUserId,
				Variables:       taskVars,
				AutoApprove:     true,
				StopOnViolation: env.StopOnViolation,
				BaseTask: models.BaseTask{
					Type: models.TaskTypeDestroy,
					// FIXME: 销毁任务应该从云模板代码库中读取 pipeline 文件
					// 或者读取环境 lastResTaskId 的 pipeline?
					Pipeline:    "",
					StepTimeout: 0,
					RunnerId:    "",
				},
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

// ===================================================================================
// 扫描任务逻辑
//

// doRunScanTask, startErr 只在任务启动出错时(执行步骤前出错)才会返回错误
func (m *TaskManager) doRunScanTask(ctx context.Context, task *models.ScanTask) (startErr error) {
	logger := m.logger.WithField("taskId", task.Id)

	changeTaskStatus := func(status, message string) error {
		if er := services.ChangeScanTaskStatus(m.db, task, status, message); er != nil {
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

	if task.Type != common.TaskTypeParse {
		if task.EnvId != "" { // 环境扫描
			if err := services.UpdateEnvModel(m.db, task.EnvId,
				models.Env{LastScanTaskId: task.Id}); err != nil {
				logger.Errorf("update env lastScanTaskId: %v", err)
				return
			}
		} else if task.TplId != "" { // 模板扫描
			if _, err := m.db.Where("id = ?", task.TplId).
				Update(&models.Template{LastScanTaskId: task.Id}); err != nil {
				logger.Errorf("update template lastScanTaskId: %v", err)
				return
			}
		}
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

		// 重新读取 task，获取最新的 containerId
		task, err = services.GetScanTaskById(m.db, task.Id)
		if err != nil {
			taskStartFailed(errors.Wrapf(err, "get task %s", task.Id.String()))
			return
		}

		runTaskReq, err := buildScanTaskReq(m.db, task, step)
		if err != nil {
			taskStartFailed(err)
			return
		}

		if err = m.runScanTaskStep(ctx, *runTaskReq, task, step); err != nil {
			logger.Infof("run task step: %v", err)
			break
		}
	}

	logger.Infof("run scan task done")
	return nil
}

func (m *TaskManager) processScanTaskDone(taskId models.Id) {
	logger := m.logger.WithField("func", "processScanTaskDone").WithField("taskId", taskId)

	dbSess := m.db

	// 重新查询获取 task，确保使用的是最新的 task 数据
	task, err := services.GetScanTaskById(dbSess, taskId)
	if err != nil {
		logger.Errorf("get task %d: %v", err)
		return
	}

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

	processTfResult := func() error {
		var (
			tsResult services.TsResult
			bs       []byte
			err      error
		)

		if task.PolicyStatus == common.PolicyStatusPassed || task.PolicyStatus == common.PolicyStatusViolated {
			if er := services.InitScanResult(dbSess, task); er != nil {
				return er
			}
			if bs, err = read(task.TfResultJsonPath()); err == nil && len(bs) > 0 {
				if tfResultJson, err := services.UnmarshalTfResultJson(bs); err == nil {
					tsResult = tfResultJson.Results
				}
			}

			if err := services.UpdateScanResult(dbSess, task, tsResult, task.PolicyStatus); err != nil {
				return fmt.Errorf("save scan result: %v", err)
			}
		}

		return err
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

	if err := updateTaskStatus(); err != nil {
		logger.Errorf("update task status error: %v", err)
	}

	if task.Type == common.TaskTypeScan {
		if err := processTfResult(); err != nil {
			logger.Errorf("process task scan: %s", err)
		}
	}
}

// buildScanTaskReq 构建扫描任务 RunTaskReq 对象
func buildScanTaskReq(dbSess *db.Session, task *models.ScanTask, step *models.TaskStep) (taskReq *runner.RunTaskReq, err error) {
	taskReq = &runner.RunTaskReq{
		RunnerId:        task.RunnerId,
		TaskId:          string(task.Id),
		Timeout:         task.StepTimeout,
		RepoAddress:     task.RepoAddr,
		RepoRevision:    task.CommitId,
		StopOnViolation: true,
		DockerImage:     task.Flow.Image,
		ContainerId:     task.ContainerId,
		//Repos: []runner.Repository{
		//	{
		//		RepoAddress:  task.RepoAddr,
		//		RepoRevision: task.CommitId,
		//	},
		//},
	}

	if step.Type == models.TaskStepOpaScan {
		taskReq.Policies, err = services.GetTaskPolicies(dbSess, task.Id)
		if err != nil {
			return nil, errors.Wrapf(err, "get scan task '%s' policies error: %v", task.Id, err)
		}
	}

	return taskReq, nil
}

func (m *TaskManager) runScanTaskStep(ctx context.Context, taskReq runner.RunTaskReq,
	task *models.ScanTask, step *models.TaskStep) (err error) {
	logger := m.logger.WithField("taskId", taskReq.TaskId)
	logger = logger.WithField("func", "runScanTaskStep").
		WithField("step", fmt.Sprintf("%d(%s)", step.Index, step.Type))

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("run task step painc: %v", r)
			logger.Errorln(err)
			logger.Debugf("%s", debug.Stack())
		}
	}()

	if step.NextStep != "" {
		if nextStep, err := services.GetTaskStepByStepId(m.db, step.NextStep); err != nil {
			err = errors.Wrapf(err, "get task step %s", string(step.NextStep))
			return err
		} else if nextStep.MustApproval {
			taskReq.PauseTask = true
		}
	}

	changeStepStatus := func(status, message string) {
		var er error
		if er = services.ChangeTaskStepStatusAndUpdate(m.db, task, step, status, message); er != nil {
			er = errors.Wrap(er, "update step status error")
			logger.Error(er)
			panic(err)
		}
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
		case models.TaskStepPending:
			// 先将步骤置为 running 状态，然后再发起调用，保证步骤不会重复执行
			changeStepStatus(models.TaskStepRunning, "")
			logger.Infof("start task step %d(%s)", step.Index, step.Type)
			if cid, _, err := StartTaskStep(taskReq, *step); err != nil {
				logger.Errorf("start task step error: %s", err.Error())
				changeStepStatus(models.TaskStepFailed, err.Error())
				return err
			} else if taskReq.ContainerId == "" {
				if err := services.UpdateTaskContainerId(m.db, models.Id(taskReq.TaskId), cid); err != nil {
					panic(errors.Wrapf(err, "update job %s container id", taskReq.TaskId))
				}
			}
		case models.TaskStepRunning:
			if _, err = WaitScanTaskStep(ctx, m.db, task, step); err != nil {
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
