package models

import (
	"cloudiac/common"
	"cloudiac/portal/libs/db"
	"cloudiac/runner"
	"cloudiac/utils"
	"database/sql/driver"
	"fmt"
	"path/filepath"
	"time"
)

type TaskBackendInfo struct {
	BackendUrl  string `json:"backend_url"`
	CtServiceId string `json:"ct_service_id"`
	LogFile     string `json:"log_file"`
	ContainerId string `json:"container_id"`
}

func (b TaskBackendInfo) Value() (driver.Value, error) {
	return MarshalValue(b)
}

func (b *TaskBackendInfo) Scan(value interface{}) error {
	return UnmarshalValue(value, b)
}

type TaskVariables []VariableBody

func (v TaskVariables) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *TaskVariables) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

type TaskResult struct {
	ResAdded     int      `json:"resAdded"`
	ResChanged   int      `json:"resChanged"`
	ResDestroyed int      `json:"resDestroyed"`
	StateResList []string `json:"stateResList"`
}

func (v TaskResult) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *TaskResult) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

type TaskExtra struct {
	Source       string `json:"source"`
	TransitionId string `json:"transitionId"`
}

func (v TaskExtra) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *TaskExtra) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

const (
	TaskTypePlan    = common.TaskTypePlan
	TaskTypeApply   = common.TaskTypeApply
	TaskTypeDestroy = common.TaskTypeDestroy

	TaskPending  = common.TaskPending
	TaskRunning  = common.TaskRunning
	TaskFailed   = common.TaskFailed
	TaskComplete = common.TaskComplete
	TaskTimeout  = common.TaskTimeout
)

var TaskStatusList = []string{TaskPending, TaskRunning, TaskFailed, TaskComplete, TaskTimeout}

type Task struct {
	SoftDeleteModel

	OrgId     Id `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id `json:"projectId" gorm:"size:32;not null"`
	TplId     Id `json:"TplId" gorm:"size:32;not null"`
	EnvId     Id `json:"envId" gorm:"size:32;not null"`

	Name      string `json:"name" gorm:"not null;comment:'任务名称'"`
	CreatorId Id     `json:"size:32;creatorId"`
	RunnerId  string `json:"runnerId" gorm:"not null"`
	CommitId  string `json:"commitId" gorm:"not null"`

	// 任务每一步的执行超时，整个任务无超时控制
	StepTimeout int `json:"stepTimeout" gorm:"default:'600';comment:'执行超时'"`

	AutoApprove bool `json:"autoApproval" gorm:"default:'0'"`

	Status  string `json:"status"`  // gorm 配置见 Migrate()
	Message string `json:"message"` // 任务的状态描述信息，如失败原因

	Type     string   `json:"type" gorm:"not null;enum('plan', 'apply', 'destroy')"`
	Flow     TaskFlow `json:"-" gorm:"type:json"`
	CurrStep int      `json:"currStep" gorm:"default:'0'"` // 当前在执行的流程步骤
	Targets  StrSlice `json:"targets" gorm:"type:json"`    // 指定 terraform target 参数

	StartAt *time.Time `json:"startAt" gorm:"null;comment:'任务开始时间'"`
	EndAt   *time.Time `json:"endAt" gorm:"null;comment:'任务结束时间'"`

	// 本次执行使用的所有变量(继承、覆盖计算之后的)
	Variables TaskVariables `json:"variables" gorm:"type:json"`

	// 任务执行结果，如 add/change/delete 的资源数量等
	Result TaskResult `json:"result" gorm:"type:json"`

	// 扩展属性，包括 source, transitionId 等
	Extra TaskExtra `json:"extra" gorm:"type:json"`
}

func (Task) TableName() string {
	return "iac_task"
}

func (t *Task) Exited() bool {
	return t.IsExitedStatus(t.Status)
}

func (t *Task) Started() bool {
	return t.IsStartedStatus(t.Status)
}

func (Task) IsStartedStatus(status string) bool {
	return !utils.InArrayStr([]string{TaskPending}, status)
}

func (Task) IsExitedStatus(status string) bool {
	return utils.InArrayStr([]string{TaskFailed, TaskComplete, TaskTimeout}, status)
}

func (t *Task) Migrate(sess *db.Session) (err error) {
	// 以下 column 通过 Migrate 来维护，确保新增加的 enum 生效
	columnDefines := []struct {
		column     string
		typeDefine string
	}{
		{
			"status",
			`ENUM('pending','running','failed','complete','timeout') DEFAULT 'pending' COMMENT '作业状态'`,
		},
	}
	for _, cd := range columnDefines {
		if err := sess.DB().ModifyColumn(cd.column, cd.typeDefine).Error; err != nil {
			return err
		}
	}

	return nil
}

type TaskStepBody struct {
	Type string   `json:"type" yaml:"type" gorm:"type:enum('init','plan','apply','play','command','destroy')"`
	Name string   `json:"name" yaml:"name" gorm:""`
	Args StrSlice `json:"args" yaml:"args" gorm:"type:text"`
}

const (
	TaskStepInit    = common.TaskStepInit
	TaskStepPlan    = common.TaskStepPlan
	TaskStepApply   = common.TaskStepApply
	TaskStepDestroy = common.TaskStepDestroy
	TaskStepPlay    = common.TaskStepPlay
	TaskStepCommand = common.TaskStepCommand

	TaskStepPending   = common.TaskStepPending
	TaskStepApproving = common.TaskStepApproving
	TaskStepRejected  = common.TaskStepRejected
	TaskStepRunning   = common.TaskStepRunning
	TaskStepFailed    = common.TaskStepFailed
	TaskStepComplete  = common.TaskStepComplete
	TaskStepTimeout   = common.TaskStepTimeout
)

type TaskStep struct {
	BaseModel
	TaskStepBody

	OrgId     Id         `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id         `json:"projectId" gorm:"size:32;not null"`
	EnvId     Id         `json:"envId" gorm:"size:32;not null"`
	TaskId    Id         `json:"taskId" gorm:"size:32;not null"`
	Index     int        `json:"index" gorm:"size:32;not null"`
	Status    string     `json:"status" gorm:"type:enum('pending','approving','rejected','running','failed','complete','timeout')"`
	Message   string     `json:"message" gorm:"type:text"`
	StartAt   *time.Time `json:"startAt"`
	EndAt     *time.Time `json:"endAt"`
	LogPath   string     `json:"logPath" gorm:""`

	ApproverId Id `json:"approverId" gorm:"size:32;not null"` // 审批者用户 id
}

func (TaskStep) TableName() string {
	return "iac_task_step"
}

func (s *TaskStep) IsApproved() bool {
	// 只有 apply 和 destroy 步骤需要审批
	if utils.StrInArray(s.Type, TaskStepApply, TaskStepDestroy) &&
		len(s.ApproverId) == 0 &&
		s.Status != TaskStepRejected {
		return false
	}
	return true
}

func (s *TaskStep) GenLogPath() string {
	return filepath.Join(
		s.ProjectId.String(),
		s.EnvId.String(),
		s.TaskId.String(),
		fmt.Sprintf("step%d", s.Index),
		runner.TaskLogName,
	)
}
