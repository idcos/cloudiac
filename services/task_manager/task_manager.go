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

	// 用于 task 管理协程通知 manager 其己退出
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

func (m *TaskManager) init() error {
	m.db = db.Get()
	m.tasks = make(map[string]*models.Task)
	m.runnerTasks = make(map[string]int)
	m.wg = sync.WaitGroup{}

	// 关闭前一次的 chan(init 可以被重复调用)
	if m.taskExitedCh != nil {
		close(m.taskExitedCh)
	}
	m.taskExitedCh = make(chan string)
	go m.listenTaskExited()

	// 初始化每个 runner 运行的任务数量
	{
		results := make([]struct {
			CtServiceId string
			Count       int
		}, 0)

		err := m.db.Model(&models.Task{}).
			Where("status = ?", consts.TaskRunning).Group("ct_service_id").
			Select("ct_service_id,COUNT(*) AS count").Scan(&results)
		if err != nil {
			m.logger.Errorf("count runner %s tasks error: %v", consts.TaskRunning, err)
			return err
		}

		for _, r := range results {
			m.runnerTasks[r.CtServiceId] = r.Count
		}

		m.maxTasksPerRunner = services.GetRunnerMax()
	}

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
	// ctx 用于
	// 	1. 通知释放锁
	//	2. task 协程退出
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		m.waitTaskRunExit()
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

	ticker := time.NewTicker(time.Millisecond * 500)
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

func (m *TaskManager) run(ctx context.Context) {
	tasks := make([]*models.Task, 0)
	maxTasksOnce := 256

	// TODO 查询时过滤掉己达并发限制的 runner
	if err := m.db.Where("status = ?", consts.TaskPending).
		Order("id").Limit(maxTasksOnce).Find(&tasks); err != nil {
		m.logger.Errorf("find '%s' tasks error: %v", consts.TaskPending, err)
		return
	}

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			break
		default:
		}

		if err := m.runTask(ctx, task); err != nil {
			runnerId := task.CtServiceId
			if err == ErrMaxTasksPerRunner {
				m.logger.WithField("count", m.runnerTasks[runnerId]).Infof("runner %s: %v", runnerId, err)
			} else {
				m.logger.WithField("taskId", task.Guid).Errorf("run task error: %v", err)
			}
		}
	}
}

// 启动任务
func (m *TaskManager) runTask(ctx context.Context, task *models.Task) error {
	// 判断并发数量
	n := m.runnerTasks[task.CtServiceId]
	if n > m.maxTasksPerRunner {
		return ErrMaxTasksPerRunner
	}

	m.wg.Add(1)
	go func() {
		defer func() {
			m.wg.Done()
			m.taskExitedCh <- task.Guid
		}()

		m.logger.Infof("start task: %v", task.Guid)
		services.StartTask(ctx, m.db, *task)
	}()

	m.tasks[task.Guid] = task
	m.runnerTasks[task.CtServiceId] += 1
	return nil
}

// 任务结束，将其从 manager 管理状态中移除
func (m *TaskManager) listenTaskExited() {
	logger := m.logger
	for {
		taskId, ok := <-m.taskExitedCh
		if !ok {
			break
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

func (m *TaskManager) waitTaskRunExit() {
	m.logger.Infof("waiting all task goroutine exit ...")
	m.wg.Wait()
	m.logger.Infof("all task goroutine exited")
}
