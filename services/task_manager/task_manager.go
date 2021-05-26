package task_manager

import (
	"cloudiac/configs"
	"fmt"
	"runtime/debug"
	"time"

	"cloudiac/consts"
	"cloudiac/libs/db"
	"cloudiac/models"
	"cloudiac/services"
	"cloudiac/utils/logs"
)

var (
	ErrMaxTasksPerRunner = fmt.Errorf("concurrent tasks is limited")
)

type TaskManager struct {
	id     string
	db     *db.Session
	logger logs.Logger

	tasks       map[string]*models.Task
	runnerTasks map[string]int // 每个 runner 正在执行的任务数量

	// 用于 task 管理协程通知 manager 其己退出
	taskExitedCh chan string

	maxTasksPerRunner int // 每个 runner 并发任务数量限制
}

func newTaskManager() *TaskManager {
	id := configs.Get().Consul.ServiceID
	return &TaskManager{
		id:     id,
		logger: logs.Get().WithField("worker", "taskManager").WithField("portalId", id),
	}
}

func Start() {
	m := newTaskManager()
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
		time.Sleep(time.Second)
	}
}

func (m *TaskManager) init() error {
	m.db = db.Get()
	m.tasks = make(map[string]*models.Task)
	m.runnerTasks = make(map[string]int)

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

func (m *TaskManager) start() {
	m.logger.Infof("task manager started")

	m.init()
	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()

	for {
		// TODO 使用分布式锁确保只有一个 manager 在运行

		select {
		case <-ticker.C:
			m.run()
		}
	}
}

func (m *TaskManager) run() {
	tasks := make([]*models.Task, 0)
	maxTasksOnce := 256

	// TODO 查询时过滤掉己达并发限制的 runner
	if err := m.db.Where("status = ?", consts.TaskPending).
		Order("id").Limit(maxTasksOnce).Find(&tasks); err != nil {
		m.logger.Errorf("find '%s' tasks error: %v", consts.TaskPending, err)
		return
	}

	for _, task := range tasks {
		if err := m.runTask(task); err != nil {
			if err == ErrMaxTasksPerRunner {
				m.logger.Debugf("runner %s: %v", task.CtServiceId, err)
			} else {
				m.logger.WithField("taskId", task.Guid).Errorf("run task error: %v", err)
			}
		}
	}
}

// 启动任务
func (m *TaskManager) runTask(task *models.Task) error {
	// 判断并发数量
	n := m.runnerTasks[task.CtServiceId]
	if n > m.maxTasksPerRunner {
		return ErrMaxTasksPerRunner
	}

	go func() {
		defer func() {
			m.taskExitedCh <- task.Guid
		}()

		m.logger.Infof("start task: %v", task.Guid)
		services.StartTask(m.db, *task)
	}()

	m.tasks[task.Guid] = task
	m.runnerTasks[task.CtServiceId] += 1
	return nil
}

// 任务结束，将其从 manager 管理状态中移除
func (m *TaskManager) listenTaskExited() {
	logger := m.logger.WithField("func", "listenTaskExited")
	for {
		taskId, ok := <-m.taskExitedCh
		if !ok {
			break
		}

		if task, ok := m.tasks[taskId]; !ok {
			logger.Warnf("unknown task '%s'", taskId)
			continue
		} else {
			m.logger.Infof("task exited: %v", taskId)
			m.runnerTasks[task.CtServiceId] -= 1
			delete(m.tasks, taskId)
		}
	}
}
