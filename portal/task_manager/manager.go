package task_manager

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/portal/services/logstorage"
	"cloudiac/portal/services/sshkey"
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

	// 通知任务开始执行
	taskStartingCh chan *models.Task
	// 通知任务己启动
	taskStartedCh chan *models.Task
	// 通知任务结束
	taskExitedCh chan string

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

	m.taskStartingCh = make(chan *models.Task)
	m.taskStartedCh = make(chan *models.Task)
	m.taskExitedCh = make(chan string)

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

	// 查询待运行任务列表的间隔
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		m.run(ctx)

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
	go func() {
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				break
			default:
				logger.Infof("recover running task %s", task.Id)
				m.runTask(ctx, task)
			}
		}
	}()

	return nil
}

func (m *TaskManager) run(ctx context.Context) {
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

	selectTable := fmt.Sprintf("iac_task, "+
		"(SELECT env_id, min(created_at) AS created_at FROM iac_task "+
		"WHERE status = '%s' GROUP BY env_id) AS min_task", models.TaskPending)
	query := m.db.Unscoped().Table(selectTable).Where(
		"min_task.env_id = iac_task.env_id AND min_task.created_at = iac_task.created_at")

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

		if v, loaded := m.envRunningTask.LoadOrStore(task.EnvId, task); loaded {
			t := v.(*models.Task)
			logger.Infof("environment '%s' has running task '%s', wait it", task.EnvId, t.Id)
			continue
		}

		m.wg.Add(1)
		go func() {
			defer func() {
				m.envRunningTask.Delete(task.EnvId)
				m.wg.Done()
			}()

			m.runTask(ctx, task)
			if task.IsEffectTask() {
				// 不管任务成功还是失败都执行
				m.processTaskDone(task)
			}
		}()
	}
}

func (m *TaskManager) runTask(ctx context.Context, task *models.Task) {
	logger := m.logger.WithField("taskId", task.Id)

	changeTaskStatus := func(status, message string) error {
		if er := services.ChangeTaskStatus(m.db, task, status, message); er != nil {
			logger.Errorf("update task status error: %v", er)
			return er
		}
		return nil
	}

	taskFailed := func(err error) {
		logger.Infoln(err)
		_ = changeTaskStatus(models.TaskFailed, err.Error())
	}

	logger.Infof("run task: %s", task.Id)
	// 先更新任务为 running 状态
	// 极端情况下任务未执行好过重复执行，所以先设置状态，后发起调用
	if err := changeTaskStatus(models.TaskRunning, ""); err != nil {
		return
	}

	if task.IsEffectTask() {
		if _, er := m.db.Model(&models.Env{}).Where("id = ?", task.EnvId).
			Update(&models.Env{LastTaskId: task.Id}); er != nil {
			logger.Errorf("update env lastTaskId: %v", er)
			return
		}
	}

	runTaskReq, err := buildRunTaskReq(*task)
	if err != nil {
		taskFailed(err)
		return
	}

	steps, err := services.GetTaskSteps(m.db, task.Id)
	if err != nil {
		taskFailed(errors.Wrap(err, "get task steps"))
		return
	}

	for _, step := range steps {
		if step.Index < task.CurrStep {
			// 跳过己执行的步骤
			continue
		}

		task.CurrStep = step.Index
		if _, err = m.db.Model(task).Update(task); err != nil {
			logger.Errorf("update task error: %v", err)
			return
		}

		if err = m.runTaskStep(ctx, *runTaskReq, task, step); err != nil {
			if err == context.Canceled {
				logger.Infof("run task step: %v", err)
				return
			}
			taskFailed(errors.Wrap(err, fmt.Sprintf("step %d", step.Index)))
			return
		}
	}

	logger.Infof("task done: %s", task.Id)
}

func (m *TaskManager) processTaskDone(task *models.Task) {
	logger := m.logger.WithField("func", "processTaskDone")

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

	// 分析环境资源、outputs, changes
	processResult := func() error {
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
		}

		if bs, err := read(task.PlanJsonPath()); err != nil {
			return fmt.Errorf("read plan json: %v", err)
		} else if len(bs) > 0 {
			tfPlan, err := services.UnmarshalPlanJson(bs)
			if err = services.SaveTaskChanges(dbSess, task, tfPlan.ResourceChanges); err != nil {
				return fmt.Errorf("save task changes: %v", err)
			}
		}
		return nil
	}
	if err := processResult(); err != nil {
		logger.Errorf("process result: %v", err)
	}

	// 设置 auto destroy
	processAutoDestroy := func() error {
		env, err := services.GetEnv(dbSess, task.EnvId)
		if err != nil {
			return errors.Wrapf(err, "get env '%s'", task.EnvId)
		}

		if env.AutoDestroyAt == nil && env.TTL != "" {
			ttl, err := time.ParseDuration(env.TTL)
			if err != nil {
				return err
			}
			at := time.Now().Add(ttl)
			env.AutoDestroyAt = &at
		}
		return nil
	}
	if err := processAutoDestroy(); err != nil {
		logger.Errorf("process auto destroy at: %v", err)
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

	var stepResult *waitStepResult
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
			changeStepStatus(models.TaskStepRunning, "")
			logger.Infof("start task step %d(%s)", step.Index, step.Type)
			if err = StartTaskStep(taskReq, *step); err != nil {
				logger.Errorf("start task step error: %s", err.Error())
				changeStepStatus(models.TaskStepFailed, err.Error())
				return err
			}
		case models.TaskStepRunning:
			if stepResult, err = WaitTaskStep(ctx, m.db, task, step); err != nil {
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
		// 有可能步骤变为 complete 状态后程序强制退出，导致任务和环境状态未设置，
		// 重启后任务会被恢复执行，此时 stepResult 为 nil
		if stepResult == nil {
			// 有可能任务状态未与步骤状态同步，这里同步一下
			if er := services.ChangeTaskStatus(m.db, task, step.Status, ""); er != nil {
				logger.Errorf("change task status: %v", er)
				panic(er)
			}
		}
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
func buildRunTaskReq(task models.Task) (taskReq *runner.RunTaskReq, err error) {
	var (
		//env        *models.Env
		privateKey []byte
	)

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

	privateKey, err = sshkey.LoadPrivateKeyPem()
	if err != nil {
		return nil, errors.Wrapf(err, "load private key")
	}

	pk, err := utils.AesEncrypt(string(privateKey))
	if err != nil {
		return nil, errors.Wrap(err, "encrypt private key")
	}

	taskReq = &runner.RunTaskReq{
		Env:          runnerEnv,
		RunnerId:     task.RunnerId,
		TaskId:       string(task.Id),
		DockerImage:  "",
		StateStore:   stateStore,
		RepoAddress:  task.RepoAddr,
		RepoRevision: task.CommitId,
		Timeout:      task.StepTimeout,
		PrivateKey:   utils.EncodeSecretVar(pk, true),
	}
	return taskReq, nil
}
