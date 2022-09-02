// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package task_manager

import (
	"cloudiac/common"
	"cloudiac/configs"
	"cloudiac/policy"
	"cloudiac/portal/apps"
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
	"os"
	"path"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	locker, err := consul.GetLocker(TaskManagerLockKey, []byte(m.id), configs.Get().Consul.Address, false)
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

	// 启动账单采集定时任务
	billCron(ctx)

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
		if err := m.processAutoDeploy(); err != nil {
			m.logger.Errorf("process auto deploy error: %v", err)
		}

		m.processPendingTask(ctx)
		// 执行所有偏移检测任务
		m.beginCronDriftTask()
		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			m.logger.Infof("context done: %v", ctx.Err())
			return
		}
	}
}

// 开始所有漂移检测任务
func (m *TaskManager) beginCronDriftTask() {
	logger := m.logger.WithField("func", "beginCronDriftTask")
	cronDriftEnvs := make([]*models.Env, 0)
	query := m.db.Where("status = ? and open_cron_drift = ? and next_drift_task_time <= ?",
		models.EnvStatusActive, true, time.Now()).
		Where("locked = ?", false)
	if err := query.Model(&models.Env{}).Find(&cronDriftEnvs); err != nil {
		logger.Error(err)
		return
	}
	// 查询出来所有需要开启偏移检测的环境任务，并且创建
	for _, env := range cronDriftEnvs {
		logger = logger.WithField("envId", env.Id)
		task, err := services.GetTaskById(m.db, env.LastTaskId)
		if err != nil {
			logger.Errorf("get task by id error: %v", err) //nolint
			continue
		}
		// 先查询这个环境有没有排队中的偏移检测任务了, 有就不创建了
		existCronPendingTask, err := services.ListPendingCronTask(m.db, env.Id)
		if err != nil {
			logger.Errorf("list pending cron task error: %v", err) //nolint
			continue
		}
		// 如果查询出来有排队或执行中的漂移检测任务，则本次跳过
		if existCronPendingTask {
			continue
		}
		// 这里每次都去解析env表保存的最新的cron 表达式
		envCronTaskType, err := apps.GetCronTaskTypeAndCheckParam(env.CronDriftExpress, env.AutoRepairDrift, env.OpenCronDrift)
		if err != nil {
			logger.Errorf("get cron task type error: %v", err) //nolint
			continue
		}
		if envCronTaskType != "" {
			attrs := models.Attrs{}
			nextTime, err := apps.ParseCronpress(env.CronDriftExpress)
			if err != nil {
				logger.Errorf("parse cron express error: %v", err) //nolint
				continue
			}
			task.Type = envCronTaskType
			task.IsDriftTask = true
			err = services.DeleteHistoryCronTask(m.db)
			// 删除历史数据失败，继续剩余流程。
			if err != nil {
				logger.Errorf("delete expired task and task step failed, error: %v", err)
			}
			_, err = services.CloneNewDriftTask(m.db, *task, env)
			if err != nil {
				logger.Errorf("clone drift task error: %v", err) //nolint
				continue
			}

			attrs["nextDriftTaskTime"] = nextTime
			_, err = services.UpdateEnv(m.db, env.Id, attrs)
			if err != nil {
				logger.Errorf("update env, error: %v", err) //nolint
				continue
			}
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

	logger.Infof("find running tasks: %d", len(tasks))
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
			if errors.Is(err, errHasRunningTask) {
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
//nolint:cyclop
func (m *TaskManager) doRunTask(ctx context.Context, task *models.Task) (startErr error) {
	logger := m.logger.WithField("taskId", task.Id)
	scanTask, _ := services.GetMirrorScanTask(m.db, task.Id)

	changeTaskStatus := func(status, message string, skipUpdateEnv bool) error {
		if er := services.ChangeTaskStatus(m.db, task, status, message, skipUpdateEnv); er != nil {
			logger.Errorf("update task status error: %v", er)
			return er
		}
		if scanTask != nil {
			if er := services.ChangeScanTaskStatus(m.db, scanTask, status, "", message); er != nil {
				logger.Errorf("update task status error: %v", er)
				return er
			}
		}
		return nil
	}

	taskStartFailed := func(err error) {
		logger.Infof("task failed: %s", err)
		startErr = err
		_ = changeTaskStatus(models.TaskFailed, err.Error(), false)
	}

	logger.Infof("run task start")

	if task.IsDriftTask {
		env, err := services.GetEnvById(m.db, task.EnvId)
		if err != nil {
			logger.Errorf("get task environment %s: %v", task.EnvId, err)
			taskStartFailed(errors.New("get task environment failed"))
			return
		} else if env.Status != models.EnvStatusActive {
			startErr = errors.New("environment is not active")
			_ = changeTaskStatus(models.TaskFailed, startErr.Error(), true)
			return
		}
		// 每次任务启动从最新的部署配置中获取配置内容
		lastResTask, err := services.GetTaskById(m.db, env.LastResTaskId)
		if err != nil {
			logger.Errorf("Get the latest configuration of the environment： %s", err)
			taskStartFailed(errors.New("get task environment failed"))
			return
		}
		attrs := models.Attrs{
			"repoAddr":   lastResTask.RepoAddr,
			"playbook":   lastResTask.Playbook,
			"workdir":    lastResTask.Workdir,
			"tfVarsFile": lastResTask.TfVarsFile,
			"commitId":   lastResTask.CommitId,
		}
		if _, err := models.UpdateAttr(db.Get(), &models.Task{},
			attrs, "id = ?", task.Id); err != nil {
			logger.Errorf("Update the latest information of the task error: %v", err)
		}

	}

	if !task.Started() { // 任务可能为己启动状态(比如异常退出后的任务恢复)，这里判断一下
		// 先更新任务为 running 状态
		// 极端情况下任务未执行好过重复执行，所以先设置状态，后发起调用
		if err := changeTaskStatus(models.TaskRunning, "", false); err != nil {
			return
		}
	}

	if task.IsEffectTask() {
		if _, er := m.db.Model(&models.Env{}).
			Where("id = ?", task.EnvId). //nolint
			Update(&models.Env{LastTaskId: task.Id}); er != nil {
			logger.Errorf("update env lastTaskId: %v", er)
			return
		}
		if scanTask != nil {
			if _, er := m.db.Model(&models.Env{}).
				Where("id = ?", task.EnvId). //nolint
				Update(&models.Env{LastScanTaskId: task.Id}); er != nil {
				logger.Errorf("update env lastTaskId: %v", er)
				return
			}
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
	var PlanIndex int
	for _, step := range steps {
		if step.PipelineStep.Type == models.TaskStepPlan {
			PlanIndex = step.Index
		}
		if step.PipelineStep.Type == models.TaskStepApply {
			if task.Source == consts.TaskSourceDriftApply {
				if bs, err := readIfExist(task.TFPlanOutputLogPath(fmt.Sprintf("step%d", PlanIndex))); err != nil {
					logger.Errorf("read plan output log: %v", err)
				} else {
					driftInfo := ParseResourceDriftInfo(bs)
					if len(driftInfo) <= 0 {
						_ = changeTaskStatus(models.TaskStepComplete, "autoDrift source nothing changed", false)
						logger.WithField("step", fmt.Sprintf("%d(%s)", step.Index, step.Name)).
							Infof("auto task drift step stop ")
						break
					}
				}
			}

			if _, er := m.db.Model(&models.Task{}).
				Where("id = ?", step.TaskId). //nolint
				Update(&models.Task{Applied: true}); er != nil {
				logger.Errorf("update task  terraformApply applied: %v", er)
			}
		}

		startErr, runErr := m.processStartStep(ctx, task, step, *runTaskReq)
		if startErr != nil {
			taskStartFailed(startErr)
			return startErr
		}
		if runErr != nil {
			logger.WithField("step", fmt.Sprintf("%d(%s)", step.Index, step.Name)).
				Warnf("run task step error: %v", runErr)
			break
		}
	}

	if err := m.runTaskStepsDoneActions(ctx, task.Id); err != nil {
		logger.Errorf("runTaskStepsDoneActions: %v", err)
	}

	logger.Infof("run task end")
	return nil
}

func (m *TaskManager) processStartStep(
	ctx context.Context,
	task *models.Task,
	step *models.TaskStep,
	req runner.RunTaskReq) (startErr error, runErr error) {
	logger := m.logger.WithField("taskId", task.Id)

	if step.Index < task.CurrStep {
		// 跳过己执行的步骤
		return nil, nil
	}

	if _, err := m.db.Model(task).UpdateAttrs(models.Attrs{"CurrStep": step.Index}); err != nil {
		// taskStartFailed(errors.Wrap(err, "update task"))
		return errors.Wrap(err, "update task"), nil
	}
	task.CurrStep = step.Index

	{
		// 获取 task 最新的 containerId
		tTask, err := services.GetTaskById(m.db, task.Id)
		if err != nil {
			// taskStartFailed(errors.Wrapf(err, "get task %s", task.Id.String()))
			return errors.Wrapf(err, "get task %s", task.Id.String()), nil
		}
		req.ContainerId = tTask.ContainerId
	}

	runErr = m.runTaskStep(ctx, req, task, step)
	if err := m.processStepDone(task, step); err != nil {
		logger.Warnf("process step done error: %v", err)
		return nil, err
	}

	if runErr != nil {
		logger.Warnf("run task step err: %v", runErr)
		if (step.Type == common.TaskStepEnvScan || step.Type == common.TaskStepOpaScan) &&
			!task.StopOnViolation {
			// 合规任务失败不影响环境部署流程
			logger.Infof("run scan task step: %v", runErr)
			return nil, nil
		}
		return nil, runErr
	}
	return nil, nil
}

func (m *TaskManager) processStepDone(task *models.Task, step *models.TaskStep) error {
	dbSess := m.db

	changePlanResult(dbSess, task, step)

	processScanResult := func() error {
		var (
			tsResult policy.TsResult
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
			if bs, er := readIfExist(task.TfResultJsonPath()); er == nil && len(bs) > 0 {
				if tfResultJson, er := policy.UnmarshalTfResultJson(bs); er == nil {
					tsResult = tfResultJson.Results
				}
			}

			if err := services.UpdateScanResult(dbSess, scanTask, tsResult, scanTask.PolicyStatus); err != nil {
				return fmt.Errorf("save scan result: %v", err)
			}
		} else if scanTask.PolicyStatus == common.PolicyStatusFailed {
			if err := services.CleanScanResult(dbSess, task); err != nil {
				return fmt.Errorf("clean scan result err: %v", err)
			}
		}

		return err
	}

	switch step.Type {
	case common.TaskStepEnvScan:
		fallthrough
	case common.TaskStepOpaScan:
		return processScanResult()
	}
	return nil
}

func changePlanResult(dbSess *db.Session, task *models.Task, step *models.TaskStep) {
	logger := logs.Get()
	if step.Type == common.TaskStepTfPlan && step.Status == models.TaskComplete {
		err := taskDoneProcessPlan(dbSess, task, true)
		if err != nil {
			logger.Errorf("process task plan: %v", err)
		}
	}
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

func (m *TaskManager) processTaskDone(taskId models.Id) { //nolint:cyclop
	logger := m.logger.WithField("func", "processTaskDone").WithField("taskId", taskId)
	logger.Debugln("start process task done")

	dbSess := m.db

	// 检查是否任务异常退出，是的话设置扫描任务状态为失败
	scanTask, _ := services.GetMirrorScanTask(dbSess, taskId)
	if scanTask != nil && scanTask.PolicyStatus == common.PolicyStatusPending {
		scanTask.PolicyStatus = common.PolicyStatusFailed
		if err := services.ChangeScanTaskStatus(dbSess,
			scanTask, common.TaskFailed, "", "scan task not run or stopped by accident"); err != nil {
			logger.Errorf("update scan task status to failed err: %v", err)
		}
	}

	// 重新查询获取 task，确保使用的是最新的 task 数据
	task, err := services.GetTaskById(dbSess, taskId)
	if err != nil {
		logger.Errorf("get task %s: %v", taskId, err)
		return
	}

	if err := StopTaskContainers(dbSess, task.Id, task.EnvId); err != nil {
		logger.Warnf("stop task container: %v", err)
	}

	lastStep, err := services.GetTaskStep(dbSess, task.Id, task.CurrStep)
	if err != nil {
		logger.Errorf("get task step(%d) error: %v", err, task.CurrStep)
		return
	}

	// 任务被审批驳回时会即时更新状态，且不会执行资源统计步骤，所以不需要后续逻辑
	if lastStep.IsRejected() {
		return
	}

	if task.IsEffectTask() {
		if err := taskDoneProcessState(dbSess, task); err != nil {
			logger.Errorf("process task state: %v", err)
		}

		// 任务执行成功才会进行 changes 统计，失败的话基于 plan 文件进行变更统计是不准确的
		// (terraform 执行 apply 失败也不会输出资源变更情况)
		if lastStep.Status == models.TaskComplete {
			if err := taskDoneProcessPlan(dbSess, task, false); err != nil {
				logger.Errorf("process task plan: %v", err)
			}
		}
	}

	if lastStep.Status == models.TaskComplete && task.IsDriftTask {
		if err := taskDoneProcessDriftTask(logger, dbSess, task); err != nil {
			logger.Errorf("process drafit task done: %v", err)
			return
		}
	}

	if err := services.ChangeTaskStatusWithStep(dbSess, task, lastStep); err != nil {
		logger.Errorf("update task status error: %v", err)
	}

	if task.IsEffectTask() {
		// 注意：环境的 lastResTaskId 必须在资源漂移信息统计后执行
		if err = services.UpdateEnvModel(dbSess, task.EnvId, models.Env{LastResTaskId: task.Id}); err != nil {
			logger.Errorf("update env lastResTaskId: %v", err)
		} else {
			if err := taskDoneProcessAutoDestroy(dbSess, task); err != nil {
				// 注意: 该步骤需要在环境状态被更新之后执行
				logger.Errorf("process auto destroy: %v", err)
			}
			if err := taskDoneProcessAutoDeploy(dbSess, task); err != nil {
				// 注意: 该步骤需要在环境状态被更新之后执行
				logger.Errorf("process auto destroy: %v", err)
			}
		}
	}
}

type changeStepStatusFunc func(status, message string, step *models.TaskStep)

func getChangeStepStatusFunc(db *db.Session, task models.Tasker, logger logs.Logger) changeStepStatusFunc {
	return func(status, message string, step *models.TaskStep) {
		var er error
		if er = services.ChangeTaskStepStatus(db, task, step, status, message); er != nil {
			er = errors.Wrap(er, "update step status error")
			logger.Error(er)
			panic(er)
		}
	}
}

func (m *TaskManager) runTaskStep(
	ctx context.Context,
	taskReq runner.RunTaskReq,
	task *models.Task,
	step *models.TaskStep) (err error) {

	logger := m.logger.WithField("taskId", taskReq.TaskId)
	logger = logger.WithField("func", "runTaskStep").
		WithField("step", fmt.Sprintf("%d(%s)", step.Index, step.Type))

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("run task step panic: %v", r)
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

	if newStep, err := waitTaskStepApprove(ctx, m.db, task, step); err != nil {
		return err
	} else {
		step = newStep
	}

	if err := waitTaskStepDone(ctx, m.db, task, step, taskReq); err != nil {
		return err
	}

	switch step.Status {
	case models.TaskStepComplete:
		return nil
	case models.TaskStepFailed:
		message := "failed"
		if step.Message != "" {
			message = step.Message
		}
		return fmt.Errorf(message)
	case models.TaskStepTimeout:
		message := "timeout"
		if step.Message != "" {
			message = step.Message
		}
		return fmt.Errorf(message)
	case models.TaskStepAborted:
		message := "aborted"
		if step.Message != "" {
			message = step.Message
		}
		return fmt.Errorf(message)
	default:
		return fmt.Errorf("unknown step status: %v", step.Status)
	}
}

func waitTaskStepApprove(ctx context.Context, db *db.Session, task *models.Task, step *models.TaskStep) (*models.TaskStep, error) {
	logger := logs.Get().
		WithField("taskId", task.Id).
		WithField("step", fmt.Sprintf("%d(%s)", step.Index, step.Type)).
		WithField("func", "waitTaskStepApprove")
	changeStepStatus := getChangeStepStatusFunc(db, task, logger)

	var (
		newStep = step
		err     error
	)
	if step.MustApproval && !step.IsApproved() {
		logger.Infof("waitting task step approve")
		changeStepStatus(models.TaskStepApproving, "", step)
		if newStep, err = WaitTaskStepApprove(ctx, db, step.TaskId, step.Index); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil, err
			}

			logger.Errorf("wait task step approve error: %v", err)
			if !errors.Is(err, ErrTaskStepRejected) && !errors.Is(err, ErrTaskStepAborted) {
				changeStepStatus(models.TaskStepFailed, err.Error(), step)
			}
			return nil, err
		}
	}
	return newStep, nil
}

//nolint:cyclop
func waitTaskStepDone(
	ctx context.Context,
	db *db.Session,
	task *models.Task,
	step *models.TaskStep,
	taskReq runner.RunTaskReq) error {

	logger := logs.Get().
		WithField("taskId", task.Id).
		WithField("step", fmt.Sprintf("%d(%s)", step.Index, step.Name)).
		WithField("func", "waitTaskStepLoop")
	changeStepStatus := getChangeStepStatusFunc(db, task, logger)

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
			changeStepStatus(models.TaskStepRunning, "", step)
			if cid, retryAble, err := StartTaskStep(taskReq, *step); err != nil {
				logger.Warnf("start task step %s(%d): %v", step.Type, step.Index, err)

				if e.Is(err, e.TaskAborted) {
					changeStepStatus(models.TaskStepAborted, err.Error(), step)
					return err
				}

				// 如果是可重试错误，并且任务设定可以重试, 则运行重试逻辑
				if retryAble && task.RetryAble {
					if step.RetryNumber > 0 && step.CurrentRetryCount < step.RetryNumber {
						// 下次重试时间为当前任务失败时间点加任务设置重试间隔时间。
						nextRetryTime := time.Now().Unix() + int64(task.RetryDelay)
						if er := services.UpdateTaskStepRetryNum(db, step.Id, step.CurrentRetryCount+1, nextRetryTime); er != nil {
							panic(errors.Wrapf(err, "update task step retry number"))
						}
						message := fmt.Sprintf("Task step start failed and try again. The current number of retries is %d", step.CurrentRetryCount)
						changeStepStatus(models.TaskStepPending, message, step)
					}
				} else {
					changeStepStatus(models.TaskStepFailed, err.Error(), step)
					return err
				}
			} else if task.ContainerId == "" {
				if err := services.UpdateTaskContainerId(db, models.Id(taskReq.TaskId), cid); err != nil {
					panic(errors.Wrapf(err, "update task %s container id", taskReq.TaskId))
				}
			}
		case models.TaskStepRunning:
			stepResult, err := WaitTaskStep(ctx, db, task, step)
			if err != nil {
				logger.Errorf("wait task result error: %v", err)
				changeStepStatus(models.TaskStepFailed, err.Error(), step)
				return err
			}
			// 合规检测步骤不通过，不需要重试，跳出循环
			if (step.Type == models.TaskStepEnvScan || step.Type == models.TaskStepOpaScan) &&
				stepResult.Result.ExitCode == common.TaskStepPolicyViolationExitCode {
				message := "Scan task step finished with violations found."
				changeStepStatus(models.TaskStepFailed, message, step)
				return nil
			}
			if stepResult.Status == models.TaskStepFailed || stepResult.Status == models.TaskStepTimeout {
				if task.RetryAble && step.RetryNumber > 0 && step.CurrentRetryCount < step.RetryNumber {
					step.NextRetryTime = time.Now().Unix() + int64(task.RetryDelay)
					step.CurrentRetryCount += 1
					message := fmt.Sprintf("Task step start failed and try again. The current number of retries is %d", step.CurrentRetryCount)
					changeStepStatus(models.TaskStepPending, message, step)
				}
			}
		default:
			return nil
		}
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

	if runnerEnv.TfVersion == "" {
		runnerEnv.TfVersion = consts.DefaultTerraformVersion
	}
	if err := buildTaskReqEnvVars(&runnerEnv, task.Variables); err != nil {
		return nil, err
	}

	stateStore := runner.StateStore{
		Backend: "consul",
		Scheme:  "http",
		Path:    task.StatePath,
		Address: "",
	}

	if configs.Get().Consul.ConsulAcl {
		stateStore.ConsulAcl = configs.Get().Consul.ConsulAcl
		stateStore.ConsulToken = configs.Get().Consul.ConsulAclToken
	}

	if configs.Get().Consul.ConsulTls {
		stateStore.ConsulTls = configs.Get().Consul.ConsulTls
		stateStore.CaPath = path.Join(common.ConsulContainerPath, common.ConsulCa)
		stateStore.CakeyPath = path.Join(common.ConsulContainerPath, common.ConsulCakey)
		stateStore.CapemPath = path.Join(common.ConsulContainerPath, common.ConsulCapem)
	}

	pk := ""
	if task.KeyId != "" {
		mKey, err := services.GetKeyById(dbSess, task.KeyId, false)
		if err != nil {
			return nil, errors.Wrapf(err, "get key '%s'", task.KeyId)
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
		RepoBranch:      task.Revision,
		RepoCommitId:    task.CommitId,
		NetworkMirror:   services.GetRegistryMirrorUrl(dbSess),
		Timeout:         task.StepTimeout,
		StopOnViolation: task.StopOnViolation,
		ContainerId:     task.ContainerId,
		CreatorId:       task.CreatorId.String(),
	}

	if err := runTaskReqAddSysEnvs(taskReq); err != nil {
		return nil, err
	}

	if scanStep, err := services.GetTaskScanStep(dbSess, task.Id); err == nil && scanStep != nil {
		policies, err := services.GetTaskPolicies(dbSess, &task)
		if err != nil {
			return nil, errors.Wrapf(err, "get task '%s' policies", task.Id)
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
		Where("locked = ?", false).
		Where("auto_destroy_task_id = ?", "").
		Where("task_status NOT IN (?)", models.EnvTaskStatus).
		Where("auto_destroy_at <= ?", time.Now()).
		Order("auto_destroy_at").Limit(limit).Find(&destroyEnvs)

	if err != nil {
		return errors.Wrapf(err, "query destroy task")
	}

	for _, env := range destroyEnvs {
		err = deployOrDestroy(env, logger, dbSess, "destroy")
		if err != nil {
			break
		}
	}

	return nil
}

func (m *TaskManager) processAutoDeploy() error {
	logger := m.logger.WithField("func", "processAutoDeploy")

	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("panic: %v", r)
			logger.Debugf("%s", debug.Stack())
		}
	}()

	dbSess := m.db
	limit := 64
	deployEnvs := make([]*models.Env, 0, limit)
	err := dbSess.Model(&models.Env{}).
		Where("status IN (?)", []string{models.EnvStatusInactive, models.EnvStatusFailed, models.EnvStatusDestroyed}).
		Where("locked = ?", false).
		Where("auto_deploy_task_id = ?", "").
		Where("task_status NOT IN (?)", models.EnvTaskStatus).
		Where("auto_deploy_at <= ?", time.Now()).
		Order("auto_deploy_at").Limit(limit).Find(&deployEnvs)

	if err != nil {
		return errors.Wrapf(err, "query auto deploy task")
	}

	for _, env := range deployEnvs {
		err = deployOrDestroy(env, logger, dbSess, "deploy")
		if err != nil {
			break
		}
	}

	return nil
}

func deployOrDestroy(env *models.Env, lg *logrus.Entry, dbSess *db.Session, op string) error {
	const (
		OpDeploy  = "deploy"
		OpDestroy = "destroy"
	)
	logger := lg.WithField("envId", env.Id)

	tx := dbSess.Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	var task *models.Task
	var er e.Error
	if op == OpDeploy {
		task, er = services.CreateAutoDeployTask(tx, env)
	} else if op == OpDestroy {
		task, er = services.CreateAutoDestroyTask(tx, env)
	} else {
		return nil
	}

	if er != nil {
		_ = tx.Rollback()
		logger.Errorf("create auto %s task: %v", op, er)
		// 创建任务失败继续处理其他任务
		return nil
	}

	var err error
	query := tx.Model(&models.Env{}).Where("id = ?", env.Id)
	if op == OpDeploy {
		_, err = query.Update(&models.Env{AutoDeployTaskId: task.Id})
	} else if op == OpDestroy {
		_, err = query.Update(&models.Env{AutoDestroyTaskId: task.Id})
	}

	if err != nil {
		_ = tx.Rollback()
		logger.Errorf("update env error: %v", err)
		return nil
	}

	if err := tx.Commit(); err != nil {
		logger.Errorf("commit error: %v", err)
		return err
	}

	logger.Infof("created auto %s task: %s", op, task.Id)
	return nil
}

// ===================================================================================
// 扫描任务逻辑
//

// doRunScanTask, startErr 只在任务启动出错时(执行步骤前出错)才会返回错误
//nolint:cyclop
func (m *TaskManager) doRunScanTask(ctx context.Context, task *models.ScanTask) (startErr error) {
	logger := m.logger.WithField("taskId", task.Id)

	changeTaskStatus := func(status, message string) error {
		if er := services.ChangeScanTaskStatus(m.db, task, status, "", message); er != nil {
			logger.Errorf("update task status error: %v", er) //nolint
			return er
		}
		return nil
	}

	taskStartFailed := func(err error) {
		logger.Infof("task failed: %s", err)
		startErr = err
		_ = changeTaskStatus(models.TaskFailed, err.Error())
		task.PolicyStatus = common.PolicyStatusFailed
		_, _ = m.db.Save(task)
	}

	logger.Infof("run scan task: %s", task.Id)

	if !task.Started() { // 任务可能为己启动状态(比如异常退出后的任务恢复)，这里判断一下
		// 先更新任务为 running 状态
		// 极端情况下任务未执行好过重复执行，所以先设置状态，后发起调用
		if err := changeTaskStatus(models.TaskRunning, ""); err != nil {
			return
		}
	}

	if task.Type == common.TaskTypeEnvScan || (task.Type == common.TaskTypeScan && task.EnvId != "") {
		if err := services.UpdateEnvModel(m.db, task.EnvId,
			models.Env{LastScanTaskId: task.Id}); err != nil {
			logger.Errorf("update env lastScanTaskId: %v", err)
			return
		}
	} else if task.Type == common.TaskTypeTplScan || (task.Type == common.TaskTypeScan && task.EnvId == "") { // 模板扫描
		if _, err := m.db.
			Where("id = ?", task.TplId). //nolint
			Update(&models.Template{LastScanTaskId: task.Id}); err != nil {
			logger.Errorf("update template lastScanTaskId: %v", err)
			return
		}
	}

	steps, err := services.GetTaskSteps(m.db, task.Id)
	if err != nil {
		taskStartFailed(errors.Wrap(err, "get task steps"))
		return
	}

	for _, step := range steps {
		startErr, runErr := m.processStartScanStep(ctx, m.db, task, step)
		if startErr != nil {
			taskStartFailed(startErr)
			return startErr
		}
		if runErr != nil {
			logger.WithField("step", fmt.Sprintf("%d(%s)", step.Index, step.Name)).
				Warnf("run task step error: %v", runErr)
			break
		}
	}

	logger.Infof("run scan task done")
	return nil
}

func (m *TaskManager) processStartScanStep(ctx context.Context, db *db.Session, task *models.ScanTask, step *models.TaskStep) (error, error) {
	logger := logs.Get()

	if step.Index < task.CurrStep {
		// 跳过己执行的步骤
		return nil, nil
	}

	var err error
	if _, err = db.Model(task).UpdateAttrs(models.Attrs{"CurrStep": step.Index}); err != nil {
		logger.Errorf("update task error: %v", err)
		return err, nil
	}

	// 重新读取 task，获取最新的 containerId
	task, err = services.GetScanTaskById(db, task.Id)
	if err != nil {
		return errors.Wrapf(err, "get task %s", task.Id.String()), nil
	}

	runTaskReq, err := buildScanTaskReq(db, task, step)
	if err != nil {
		return err, nil
	}

	if err = m.runScanTaskStep(ctx, *runTaskReq, task, step); err != nil {
		logger.Infof("run task step: %v", err)
		return nil, err
	}
	return nil, nil
}

func (m *TaskManager) processScanTaskDone(taskId models.Id) {
	logger := m.logger.WithField("func", "processScanTaskDone").WithField("taskId", taskId)

	dbSess := m.db

	// 重新查询获取 task，确保使用的是最新的 task 数据
	task, err := services.GetScanTaskById(dbSess, taskId)
	if err != nil {
		logger.Errorf("get task %s: %v", taskId, err)
		return
	}

	if err := StopScanTaskContainers(dbSess, task.Id, task.EnvId); err != nil {
		logger.Warnf("stop task container: %v", err)
	}

	lastStep, err := services.GetTaskStep(dbSess, task.Id, task.CurrStep)
	if err != nil {
		logger.Errorf("get task step(%d) error: %v", err, task.CurrStep)
		return
	}

	// 基于最后一个步骤更新任务状态
	if err = services.ChangeTaskStatusWithStep(dbSess, task, lastStep); err != nil {
		logger.Errorf("update task status error: %v", err)
	}

	if task.Type == common.TaskTypeEnvScan || task.Type == common.TaskTypeScan || task.Type == common.TaskTypeTplScan {
		if err := sacnTaskDoneProcessTfResult(dbSess, task); err != nil {
			logger.Errorf("process task scan: %s", err)
		}
	}
}

func buildTaskReqEnvVars(env *runner.TaskEnv, variables models.TaskVariables) error {
	for _, v := range variables {
		value := v.Value
		// 旧版本创建的敏感变量保存时不会添加 secret 前缀，这里判断一下，如果敏感变量无前缀则添加
		if v.Sensitive && !strings.HasPrefix(v.Value, utils.SecretValuePrefix) {
			value = utils.EncodeSecretVar(v.Value, v.Sensitive)
		}
		switch v.Type {
		case consts.VarTypeEnv:
			env.EnvironmentVars[v.Name] = value
		case consts.VarTypeTerraform:
			env.TerraformVars[v.Name] = value
		case consts.VarTypeAnsible:
			env.AnsibleVars[v.Name] = value
		default:
			return fmt.Errorf("unknown variable type: %s", v.Type)
		}
	}
	return nil
}

// buildScanTaskReq 构建扫描任务 RunTaskReq 对象
func buildScanTaskReq(dbSess *db.Session, task *models.ScanTask, step *models.TaskStep) (taskReq *runner.RunTaskReq, err error) {
	taskReq = &runner.RunTaskReq{
		RunnerId:        task.RunnerId,
		TaskId:          string(task.Id),
		Timeout:         task.StepTimeout,
		RepoAddress:     task.RepoAddr,
		RepoBranch:      task.Revision,
		RepoCommitId:    task.CommitId,
		NetworkMirror:   services.GetRegistryMirrorUrl(dbSess),
		StopOnViolation: true,
		DockerImage:     task.Flow.Image,
		ContainerId:     task.ContainerId,
		CreatorId:       task.CreatorId.String(),
	}

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
	if runnerEnv.TfVersion == "" {
		runnerEnv.TfVersion = consts.DefaultTerraformVersion
	}
	if err := buildTaskReqEnvVars(&runnerEnv, task.Variables); err != nil {
		return nil, err
	}
	taskReq.Env = runnerEnv

	if task.Type == common.TaskTypeEnvScan || task.Type == common.TaskTypeEnvParse {
		env, _ := services.GetEnvById(dbSess, task.EnvId)

		stateStore := runner.StateStore{
			Backend: "consul",
			Scheme:  "http",
			Path:    env.StatePath,
			Address: "",
		}

		if configs.Get().Consul.ConsulAcl {
			stateStore.ConsulAcl = configs.Get().Consul.ConsulAcl
			stateStore.ConsulToken = configs.Get().Consul.ConsulAclToken
		}

		if configs.Get().Consul.ConsulTls {
			stateStore.ConsulTls = configs.Get().Consul.ConsulTls
			stateStore.CaPath = path.Join(common.ConsulContainerPath, common.ConsulCa)
			stateStore.CakeyPath = path.Join(common.ConsulContainerPath, common.ConsulCakey)
			stateStore.CapemPath = path.Join(common.ConsulContainerPath, common.ConsulCapem)
		}

		taskReq.StateStore = stateStore
	}

	if err := runTaskReqAddSysEnvs(taskReq); err != nil {
		return nil, err
	}

	if step.Type == models.TaskStepOpaScan || step.Type == models.TaskStepEnvScan || step.Type == models.TaskStepTplScan {
		taskReq.Policies, err = services.GetTaskPolicies(dbSess, task)
		if err != nil {
			return nil, errors.Wrapf(err, "get scan task '%s' policies", task.Id)
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
			err = fmt.Errorf("run task step panic: %v", r)
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

	if err := waitScanTaskStepDone(ctx, m.db, task, step, taskReq); err != nil {
		return err
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
	case models.TaskStepAborted:
		return errors.New("aborted")
	default:
		return fmt.Errorf("unknown step status: %v", step.Status)
	}
}

func waitScanTaskStepDone(
	ctx context.Context,
	db *db.Session,
	task *models.ScanTask,
	step *models.TaskStep,
	taskReq runner.RunTaskReq) error {

	logger := logs.Get().
		WithField("taskId", task.Id).
		WithField("step", fmt.Sprintf("%d(%s)", step.Index, step.Name)).
		WithField("func", "waitTaskStepApprove")

	changeStepStatus := getChangeStepStatusFunc(db, task, logger)

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
			changeStepStatus(models.TaskStepRunning, "", step)
			logger.Infof("start task step %d(%s)", step.Index, step.Type)
			if cid, _, err := StartTaskStep(taskReq, *step); err != nil {
				logger.Errorf("start task step error: %s", err.Error())

				if e.Is(err, e.TaskAborted) {
					changeStepStatus(models.TaskStepAborted, err.Error(), step)
				} else {
					changeStepStatus(models.TaskStepFailed, err.Error(), step)
				}
				return err
			} else if task.ContainerId == "" {
				if err := services.UpdateScanTaskContainerId(db, models.Id(taskReq.TaskId), cid); err != nil {
					panic(errors.Wrapf(err, "update job %s container id", taskReq.TaskId))
				}
			}
		case models.TaskStepRunning:
			if _, err := WaitScanTaskStep(ctx, db, task, step); err != nil {
				logger.Errorf("wait scan task result error: %v", err)
				changeStepStatus(models.TaskStepFailed, err.Error(), step)
				return err
			}
		default:
			return nil
		}
	}
}

// 为 req 添加 sysEnvs(直接修改传入的 req)
func runTaskReqAddSysEnvs(req *runner.RunTaskReq) error {
	sysEnvs := make(map[string]string)

	// CLOUDIAC_TASK_ID	当前任务的 id
	sysEnvs["CLOUDIAC_TASK_ID"] = req.TaskId
	sysEnvs["TF_VAR_cloudiac_task_id"] = req.TaskId
	// CLOUDIAC_BRANCH	当前任务的云模板代码的分支
	sysEnvs["CLOUDIAC_BRANCH"] = req.RepoBranch
	sysEnvs["TF_VAR_cloudiac_branch"] = req.RepoBranch
	// CLOUDIAC_COMMIT	当前任务的云模板代码 commit hash
	sysEnvs["CLOUDIAC_COMMIT"] = req.RepoCommitId
	sysEnvs["TF_VAR_cloudiac_commit"] = req.RepoCommitId

	if req.CreatorId != "" {
		user, err := services.GetUserByIdRaw(db.Get(), models.Id(req.CreatorId))
		if err != nil {
			return errors.Wrapf(err, "query user %s", req.CreatorId)
		}
		if user.Name == "" {
			sysEnvs["CLOUDIAC_USERNAME"] = user.Email
			sysEnvs["TF_VAR_cloudiac_username"] = user.Email
		} else {
			sysEnvs["CLOUDIAC_USERNAME"] = user.Name
			sysEnvs["TF_VAR_cloudiac_username"] = user.Name
		}
	}

	if req.Env.Id != "" {
		env, err := services.GetEnvById(db.Get(), models.Id(req.Env.Id))
		if err != nil {
			return errors.Wrapf(err, "query env %s", req.Env.Id)
		}

		resCount, err := services.GetEnvResourceCount(db.Get(), env.Id)
		if err != nil {
			return errors.Wrapf(err, "%s, query environment resource count", req.Env.Id)
		}

		// 当前任务的组织 ID
		sysEnvs["CLOUDIAC_ORG_ID"] = env.OrgId.String()
		sysEnvs["TF_VAR_cloudiac_org_id"] = env.OrgId.String()
		// 当前任务的项目 ID
		sysEnvs["CLOUDIAC_PROJECT_ID"] = env.ProjectId.String()
		sysEnvs["TF_VAR_cloudiac_project_id"] = env.ProjectId.String()

		// CLOUDIAC_TEMPLATE_ID	当前任务的模板 ID
		sysEnvs["CLOUDIAC_TEMPLATE_ID"] = env.TplId.String()
		sysEnvs["TF_VAR_cloudiac_template_id"] = env.TplId.String()
		// CLOUDIAC_ENV_ID	当前任务的环境 ID
		sysEnvs["CLOUDIAC_ENV_ID"] = env.Id.String()
		sysEnvs["TF_VAR_cloudiac_env_id"] = env.Id.String()
		// CLOUDIAC_ENV_NAME	当前任务的环境名称
		sysEnvs["CLOUDIAC_ENV_NAME"] = env.Name
		sysEnvs["TF_VAR_cloudiac_env_name"] = env.Name
		// CLOUDIAC_ENV_STATUS	当前环境状态(启动任务时)
		sysEnvs["CLOUDIAC_ENV_STATUS"] = env.Status
		sysEnvs["TF_VAR_cloudiac_env_status"] = env.Status
		// 当前环境中的资源数量(启动任务时)
		sysEnvs["CLOUDIAC_ENV_RESOURCES"] = fmt.Sprintf("%d", resCount)
		sysEnvs["TF_VAR_cloudiac_env_resources"] = fmt.Sprintf("%d", resCount)
		// CLOUDIAC_TF_VERSION	当前任务使用的 terraform 版本号(eg. 0.14.11)
		sysEnvs["CLOUDIAC_TF_VERSION"] = req.Env.TfVersion
		sysEnvs["TF_VAR_cloudiac_tf_version"] = req.Env.TfVersion
	}

	req.SysEnvironments = sysEnvs

	return nil
}

func ParseResourceDriftInfo(bs []byte) map[string]models.ResourceDrift {
	logger := logs.Get().WithField("func", "ParseResourceDriftInfo")
	defer func() {
		err := recover()
		if err != nil {
			logger.Errorf("parse resource drift info error: %v", err)
		}
	}()
	content := strings.Split(string(bs), "\n")
	cronTaskInfoMap := make(map[string]models.ResourceDrift)
	for k, v := range content {
		if strings.Contains(v, "#") && strings.Contains(v, "must be") || strings.Contains(v, "will be") {
			var resourceDetail string
			cronTaskInfo := models.ResourceDrift{}
			reg1 := regexp.MustCompile(`#\s\S*`)
			result1 := reg1.FindAllStringSubmatch(v, 1)
			if len(result1) == 0 {
				continue
			}
			address := stripansi.Strip(strings.TrimSpace(result1[0][0][1:]))
			for k1, v2 := range content[k+1:] {
				if ((strings.Contains(v2, "#") && strings.Contains(v2, "must be")) || strings.Contains(v2, "will be")) ||
					strings.Contains(v2, "Plan:") {
					resourceDetail = strings.Join(content[k+1:k1+k], "\n")
					break
				}
			}
			cronTaskInfo.DriftDetail = resourceDetail
			cronTaskInfoMap[address] = cronTaskInfo
		}
	}
	return cronTaskInfoMap
}
