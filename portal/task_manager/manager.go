package task_manager

import (
	"cloudiac/configs"
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/services"
	"cloudiac/portal/services/sshkey"
	"cloudiac/runner"
	"cloudiac/utils"
	"cloudiac/utils/consul"
	"cloudiac/utils/logs"
	"context"
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"net/url"
	"runtime/debug"
	"strings"
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

	tasks       map[string]*models.Task
	runnerTasks map[string]int // 每个 runner 正在执行的任务数量

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
	m.tasks = make(map[string]*models.Task)
	m.runnerTasks = make(map[string]int)
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
	go func() {
		<-ctx.Done()
		if err := locker.Unlock(); err != nil {
			if err != consulapi.ErrLockNotHeld {
				m.logger.Errorf("release lock error: %v", err)
			}
		}
		close(stopLockCh)
	}()

	lockLostCh, err := locker.Lock(stopLockCh)
	if err != nil {
		return nil, errors.Wrap(err, "acquire lock")
	}

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
	ticker := time.NewTicker(time.Second)
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
		Where("status = ?", models.TaskRunning)

	tasks := make([]*models.Task, 0)
	if err := query.Find(&tasks); err != nil {
		logger.Errorf("find %s tasks error: %v", models.TaskRunning, err)
		return err
	}

	logger.Infof("find %d running tasks", len(tasks))
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
	for runnerId, count := range m.runnerTasks {
		if count >= m.maxTasksPerRunner {
			limitedRunners = append(limitedRunners, runnerId)
		}
	}

	queryTaskLimit := 64 // 单次查询任务数量限制
	query := m.db.Model(&models.Task{}).
		Where("status = ?", models.TaskPending).
		Order("id").
		Limit(queryTaskLimit)

	// 查询时过滤掉己达并发限制的 runner
	if len(limitedRunners) > 0 {
		query = query.Where("ct_service_id NOT IN (?)", limitedRunners)
	}

	tasks := make([]*models.Task, 0)
	if err := query.Find(&tasks); err != nil {
		logger.Panicf("find '%s' tasks error: %v", models.TaskPending, err)
	}

	for i := range tasks {
		select {
		case <-ctx.Done():
			break
		default:
		}

		task := tasks[i]
		// 判断 runner 并发数量
		n := m.runnerTasks[task.RunnerId]
		if n >= m.maxTasksPerRunner {
			logger.WithField("count", n).Infof("runner %s: %v", task.RunnerId, ErrMaxTasksPerRunner)
			continue
		}

		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.runTask(ctx, task)
		}()
	}
}

func (m *TaskManager) runTask(ctx context.Context, task *models.Task) {
	logger := m.logger.WithField("taskId", task.Id)
	logger.Infof("run task %s", task.Id)

	updateTask := func() error {
		if _, err := m.db.Model(task).Update(task); err != nil {
			logger.Errorf("update task error: %v", err)
			return err
		}
		return nil
	}

	taskFailed := func(err error) {
		logger.Infoln(err)
		now := time.Now()
		task.Status = models.TaskFailed
		task.Message = err.Error()
		task.EndAt = &now
		_ = updateTask()
	}

	// 先更新任务为 running 状态
	// 极端情况下任务未执行好过重复执行，所以先设置状态，后发起调用
	task.Status = models.TaskRunning
	task.Message = ""
	now := time.Now()
	task.StartAt = &now
	if err := updateTask(); err != nil {
		return
	}

	runTaskReq, err := buildRunTaskReq(m.db, *task)
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
		if err = updateTask(); err != nil {
			return
		}

		if err = m.runTaskStep(ctx, *runTaskReq, task, step); err != nil {
			taskFailed(errors.Wrap(err, fmt.Sprintf("run task step %d", step.Index)))
			return
		}
	}

	task.Status = models.TaskStepComplete
	now = time.Now()
	task.EndAt = &now
	_ = updateTask()
}

func (m *TaskManager) runTaskStep(ctx context.Context, taskReq runner.RunTaskReq,
	task *models.Task, step *models.TaskStep) (err error) {
	logger := m.logger.WithField("taskId", taskReq.TaskId)
	logger = logger.WithField("func", "runTaskStep").WithField("step", fmt.Sprintf("%d", step.Index))

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("painc: %v", r)
		}
	}()

	updateStep := func() {
		if _, er := m.db.Model(step).Update(step); er != nil {
			er = errors.Wrapf(er, "update step error")
			logger.Error(er)
			panic(er)
		}
	}

	if !task.AutoApprove && !step.IsApproved() {
		logger.Infof("waitting task step approve")

		step.Status = models.TaskStepApproving
		step.Message = ""
		updateStep()

		var newStep *models.TaskStep
		if newStep, err = WaitTaskStepApprove(ctx, m.db, step.TaskId, step.Index); err != nil {
			logger.Errorf("wait task step approve error: %v", err)
			if err == ErrTaskStepRejected {
				step.Status = models.TaskStepRejected
			} else {
				step.Status = models.TaskStepFailed
			}
			step.Message = err.Error()
			updateStep()
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
			now := time.Now()
			step.Status = models.TaskStepRunning
			step.Message = ""
			step.StartAt = &now
			updateStep()

			logger.Infof("start task step %d", step.Index)
			if err = StartTaskStep(taskReq, *step); err != nil {
				logger.Errorf("start task step error: %s", err.Error())
				step.Status = models.TaskStepFailed
				step.Message = err.Error()
				updateStep()
				return err
			}
		case models.TaskStepRunning:
			if stepResult, err = WaitTaskStep(ctx, m.db, task, step); err != nil {
				logger.Errorf("wait task result error: %v", err)
				step.Status = models.TaskStepFailed
				step.Message = err.Error()
				updateStep()
				return err
			}
		default:
			break loop
		}
	}

	switch step.Status {
	case models.TaskStepComplete:
		if len(stepResult.Result.StateListContent) > 0 {
			task.Result.StateResList = strings.Split(string(stepResult.Result.StateListContent), "\n")
		}
		if stepResult.Status == models.TaskComplete {
			////TODO 解析日志输出，更新资源变更信息到 task.Result
			//tfInfo := ParseTfOutput(task.BackendInfo.LogFile)
			//models.UpdateAttr(dbSess.Where("id = ?", task.Id), &models.Task{}, tfInfo)
		}
		if _, err = m.db.Model(&models.Task{}).Update(task); err != nil {
			return
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
func buildRunTaskReq(dbSess *db.Session, task models.Task) (taskReq *runner.RunTaskReq, err error) {
	var (
		env        *models.Env
		tpl        *models.Template
		privateKey []byte
	)

	env, err = services.GetEnv(dbSess, task.EnvId)
	if err != nil {
		return nil, errors.Wrapf(err, "get env %v", task.EnvId)
	}

	tpl, err = services.GetTemplate(dbSess, env.TplId)
	if err != nil {
		return nil, errors.Wrapf(err, "get template %v", env.TplId)
	}

	runnerEnv := runner.TaskEnv{
		Id:              string(env.Id),
		Workdir:         tpl.Workdir,
		TfVarsFile:      tpl.TfVarsFile,
		Playbook:        tpl.Playbook,
		PlayVarsFile:    tpl.PlayVarsFile,
		EnvironmentVars: make(map[string]string),
		TerraformVars:   make(map[string]string),
		AnsibleVars:     make(map[string]string),
	}

	getVarValue := func(v models.VariableBody) string {
		if v.Sensitive {
			return utils.AesDecrypt(v.Value)
		}
		return v.Value
	}

	for _, v := range task.Variables {
		value := getVarValue(v)
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

	if env.TfVarsFile != "" {
		runnerEnv.TfVarsFile = env.TfVarsFile
	}
	if env.PlayVarsFile != "" {
		runnerEnv.PlayVarsFile = env.PlayVarsFile
	}

	stateStore := runner.StateStore{
		Backend: "consul",
		Scheme:  "http",
		Path:    env.StatePath,
		Address: configs.Get().Consul.Address,
	}

	privateKey, err = sshkey.LoadPrivateKeyPem()
	if err != nil {
		return nil, errors.Wrapf(err, "load private key")
	}

	var (
		repoAddr     = tpl.RepoAddr
		repoToken    = tpl.RepoToken
		repoRevision = task.CommitId
	)

	if repoRevision == "" {
		repoRevision = tpl.RepoRevision
	}

	if (repoToken == "" || !strings.Contains("://", repoAddr)) && tpl.VcsId != "" {
		var vcs *models.Vcs
		if vcs, err = services.QueryVcsByVcsId(tpl.VcsId, dbSess); err != nil {
			return nil, errors.Wrapf(err, "get vcs %s", tpl.VcsId)
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

	if u, err := url.Parse(repoAddr); err != nil {
		return nil, errors.Wrapf(err, "parse url: %v", repoAddr)
	} else if repoToken != "" {
		u.User = url.UserPassword("token", repoToken)
		repoAddr = u.String()
	}

	taskReq = &runner.RunTaskReq{
		Env:          runnerEnv,
		RunnerId:     env.RunnerId,
		TaskId:       string(task.Id),
		DockerImage:  "",
		StateStore:   stateStore,
		RepoAddress:  repoAddr,
		RepoRevision: tpl.RepoRevision,
		Timeout:      env.Timeout,
		PrivateKey:   string(privateKey),
	}
	return taskReq, nil
}
