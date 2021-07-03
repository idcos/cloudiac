package models

import (
	"cloudiac/portal/libs/db"
	"cloudiac/utils"
	"database/sql/driver"
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
	ResAdded     int `json:"resAdded"`
	ResChanged   int `json:"resChanged"`
	ResDestroyed int `json:"resDestroyed"`
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
	TaskTypePlan    = "plan"    // 计划执行，不会修改资源或做服务配置
	TaskTypeApply   = "apply"   // 执行 terraform apply 和 playbook
	TaskTypeDestroy = "destroy" // 销毁，删除所有资源

	TaskPending  = "pending"
	TaskRunning  = "running"
	TaskFailed   = "failed"
	TaskComplete = "complete"
	TaskTimeout  = "timeout"
)

var TaskStatusList = []string{TaskPending, TaskRunning, TaskFailed, TaskComplete, TaskTimeout}

type Task struct {
	SoftDeleteModel

	OrgId     Id `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id `json:"projectId" gorm:"size:32;not null"`
	EnvId     Id `json:"envId" gorm:"size:32;not null"`

	Name      string `json:"name" gorm:"not null;comment:'任务名称'"`
	CreatorId Id     `json:"size:32;creatorId"`
	RunnerId  string `json:"runnerId" gorm:"not null"`
	CommitId  string `json:"commitId" gorm:"not null"`
	Status    string `json:"status"`  // gorm 配置见 Migrate()
	Message   string `json:"message"` // 任务的状态描述信息，如失败原因

	Type     string `json:"type" gorm:"not null;enum('plan', 'apply', 'destroy')"`
	Flow     string `json:"-" gorm:"type:text"`
	CurrStep int    `json:"currStep" gorm:"default:'0'"` // 当前在执行的流程步骤

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

const (
	TaskStepInit  = "init"
	TaskStepPlan  = "plan"
	TaskStepApply = "apply"
	TaskStepPlay  = "play" // play playbook

	TaskStepPending  = "pending"
	TaskStepRunning  = "running"
	TaskStepFailed   = "failed"
	TaskStepComplete = "complete"
)

type TaskStep struct {
	BaseModel
	OrgId     Id     `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id     `json:"projectId" gorm:"size:32;not null"`
	TaskId    Id     `json:"taskId" gorm:"size:32;not null"`
	Index     Id     `json:"index" gorm:"size:32;not null"`
	Type      string `json:"type" gorm:"type:enum('init', 'plan', 'apply', 'play')"`
	Status    string `json:"status" gorm:"type:enum('pending','running','failed','complete')"`
	LogPath   string `json:"logPath" gorm:""`
}

func (TaskStep) TableName() string {
	return "iac_task_step"
}
