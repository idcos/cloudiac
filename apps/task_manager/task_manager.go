package task_manager

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"

	"cloudiac/configs"
	"cloudiac/consts"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/services"
	"cloudiac/utils/consul"
	"cloudiac/utils/logs"
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

func Start(id string) {
	if id == "" {
		id = configs.Get().Consul.ServiceID
	}
	m := TaskManager{
		id:     id,
		logger: logs.Get().WithField("worker", "taskManager").WithField("portalId", id),
	}

	recoveredRun := func() {
		defer func() {
			if r := recover(); r != nil {
				if r != nil {
					m.logger.Errorf("panic: %v", r)
					m.logger.Debugf("stack: %s", debug.Stack())
				}
			}
		}()

		m.start()
	}

	for {
		recoveredRun()
		time.Sleep(time.Second * 10)
	}
}

// 该函数会被重复调用
func (m *TaskManager) init() error {
	m.db = db.Get()
	m.tasks = make(map[string]*models.Task)
	m.runnerTasks = make(map[string]int)
	m.wg = sync.WaitGroup{}

	m.taskStartingCh = make(chan *models.Task)
	m.taskStartedCh = make(chan *models.Task)
	m.taskExitedCh = make(chan string)

	m.maxTasksPerRunner = services.GetRunnerMax()
	return nil
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
	m.init()

	go func() {
		<-lockLostCh
		m.logger.Warnf("task manager lock lost")
		cancel()
	}()

	go m.listenTaskStarting()
	go m.listenTaskStarted(ctx)
	go m.listenTaskExited()

	// 恢复执行中的任务状态
	go m.recoverTask(ctx)

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

// run 查询数据库中待执行的任务，并启动
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
		Where("status = ?", consts.TaskPending).
		Order("id").
		Limit(queryTaskLimit)

	// 查询时过滤掉己达并发限制的 runner
	if len(limitedRunners) > 0 {
		query = query.Where("ct_service_id NOT IN (?)", limitedRunners)
	}

	tasks := make([]*models.Task, 0)
	if err := query.Find(&tasks); err != nil {
		logger.Panicf("find '%s' tasks error: %v", consts.TaskPending, err)
	}

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			break
		default:
		}

		if err := m.runStartTask(ctx, task); err != nil {
			runnerId := task.CtServiceId
			if err == ErrMaxTasksPerRunner {
				logger.WithField("count", m.runnerTasks[runnerId]).Infof("runner %s: %v", runnerId, err)
			} else {
				logger.WithField("taskId", task.Guid).Errorf("run task error: %v", err)
			}
		}
	}
}

// recoverTask 查询 db 中 running 状态的任务，并通知任务 started
func (m *TaskManager) recoverTask(ctx context.Context) {
	logger := m.logger
	query := m.db.Model(&models.Task{}).
		Where("status = ?", consts.TaskRunning)

	tasks := make([]*models.Task, 0)
	if err := query.Find(&tasks); err != nil {
		logger.Panicf("find %s tasks error: %v", consts.TaskRunning, err)
	}
	logger.Infof("find %d running tasks", len(tasks))

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			break
		default:
			logger.Infof("recover running task %s", task.Guid)
			m.notifyTaskStarting(task)
			m.notifyTaskStarted(task)
		}
	}
}

// 启动任务
func (m *TaskManager) runStartTask(ctx context.Context, task *models.Task) error {
	m.logger.Infof("run task %s", task.Guid)

	// 判断 runner 并发数量
	n := m.runnerTasks[task.CtServiceId]
	if n >= m.maxTasksPerRunner {
		return ErrMaxTasksPerRunner
	}

	updateTaskStatus := func(status string) error {
		task.Status = status
		if _, err := m.db.Model(task).Update(task); err != nil {
			m.logger.Errorf("update task status(%s) error: %v", consts.TaskAssigning, err)
			return err
		}
		return nil
	}

	// 先更新任务为 assigning 状态
	// 极端情况下任务未执行好过重复执行，所以先设置状态，后发起调用
	if err := updateTaskStatus(consts.TaskAssigning); err != nil {
		return err
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		m.notifyTaskStarting(task)
		_, err := services.StartTask(m.db, task)
		if err != nil {
			m.logger.Errorf("start task error: %v", err)
			if task.Status == consts.TaskAssigning {
				// 下发失败，还原为 pending 状态，等待下次重试
				_ = updateTaskStatus(consts.TaskPending)
			}
			m.notifyTaskExited(task.Guid)
			return
		}
		m.notifyTaskStarted(task)
	}()

	return nil
}

func (m *TaskManager) notifyTaskStarting(task *models.Task) {
	m.taskStartingCh <- task
}

func (m *TaskManager) listenTaskStarting() {
	for {
		select {
		case task, ok := <-m.taskStartingCh:
			if !ok {
				return
			}
			m.tasks[task.Guid] = task
			m.runnerTasks[task.CtServiceId] += 1
		}
	}
}

func (m *TaskManager) notifyTaskStarted(task *models.Task) {
	m.taskStartedCh <- task
}

func (m *TaskManager) listenTaskStarted(ctx context.Context) {
	for {
		select {
		case task, ok := <-m.taskStartedCh:
			if !ok {
				return
			}
			m.runWaitTask(ctx, task)
		}
	}
}

func (m *TaskManager) runWaitTask(ctx context.Context, task *models.Task) {
	logger := m.logger.WithField("func", "runWaitTask")

	deadline, err := services.TaskDeadline(m.db, task.Guid)
	if err != nil {
		logger.Errorf("get task deadline error: %v", err)
		return
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		if _, err := services.WaitTaskResult(ctx, m.db, task, deadline); err != nil {
			logger.Errorf("wait task result error: %v", err)
			return
		}
		m.notifyTaskExited(task.Guid)
	}()
}

func (m *TaskManager) notifyTaskExited(taskId string) {
	m.taskExitedCh <- taskId
}

// 等待任务结束，将其从 manager 管理状态中移除
func (m *TaskManager) listenTaskExited() {
	logger := m.logger.WithField("func", "listenTaskExited")
	for {
		select {
		case taskId, ok := <-m.taskExitedCh:
			if !ok {
				return
			}

			if task, ok := m.tasks[taskId]; !ok {
				logger.Warnf("unknown task '%s'", taskId)
				continue
			} else {
				logger.WithField("taskId", taskId).Infof("task exited")
				m.runnerTasks[task.CtServiceId] -= 1
				delete(m.tasks, taskId)
			}
		}
	}
}

func (m *TaskManager) stop() {
	logger := m.logger
	logger.Infof("task manager stopping ...")

	logger.Debugf("waiting all task goroutine exit ...")
	m.wg.Wait()

	close(m.taskStartingCh)
	close(m.taskStartedCh)
	close(m.taskExitedCh)

	logger.Infof("task manager stopped")
}
