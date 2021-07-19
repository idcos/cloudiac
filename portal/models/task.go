package models

import (
	"cloudiac/common"
	"cloudiac/portal/libs/db"
	"cloudiac/runner"
	"cloudiac/utils"
	"database/sql/driver"
	"fmt"
	"path"
	"time"
)

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

	Outputs map[string]interface{} `json:"outputs"`
}

func (v TaskResult) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *TaskResult) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

type TaskExtra struct {
	Source       string `json:"source,omitempty"`
	TransitionId string `json:"transitionId,omitempty"`
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

	TaskPending   = common.TaskPending
	TaskRunning   = common.TaskRunning
	TaskApproving = common.TaskApproving
	TaskFailed    = common.TaskFailed
	TaskComplete  = common.TaskComplete
	//TaskTimeout   = common.TaskTimeout
)

type Task struct {
	SoftDeleteModel

	OrgId     Id `json:"orgId" gorm:"size:32;not null"`     // 组织ID
	ProjectId Id `json:"projectId" gorm:"size:32;not null"` // 项目ID
	TplId     Id `json:"TplId" gorm:"size:32;not null"`     // 模板ID
	EnvId     Id `json:"envId" gorm:"size:32;not null"`     // 环境ID

	Name      string `json:"name" gorm:"not null;comment:'任务名称'"` // 任务名称
	CreatorId Id     `json:"creatorId" gorm:"size:32;not null"`   // 创建人ID

	Type string `json:"type" gorm:"not null;enum('plan', 'apply', 'destroy')" enums:"'plan', 'apply', 'destroy'"` // 任务类型。1. plan: 计划 2. apply: 部署 3. destroy: 销毁

	RepoAddr string `json:"repoAddr" gorm:"not null"`
	CommitId string `json:"commitId" gorm:"not null"` // 创建任务时 revision 对应的 commit id

	Workdir      string   `json:"workdir" gorm:"default:''"`
	Playbook     string   `json:"playbook" gorm:"default:''"`
	TfVarsFile   string   `json:"tfVarsFile" gorm:"default:''"`
	PlayVarsFile string   `json:"playVarsFile" gorm:"default:''"`
	Targets      StrSlice `json:"targets" gorm:"type:json"` // 指定 terraform target 参数

	Variables TaskVariables `json:"variables" gorm:"type:json"` // 本次执行使用的所有变量(继承、覆盖计算之后的)

	StatePath string `json:"statePath" gorm:"not null"`

	// 扩展属性，包括 source, transitionId 等
	Extra TaskExtra `json:"extra" gorm:"type:json"` // 扩展属性

	KeyId       Id     `json:"keyId" gorm:"size32"`      // 部署密钥ID
	RunnerId    string `json:"runnerId" gorm:"not null"` // 部署通道
	AutoApprove bool   `json:"autoApproval" gorm:"default:'0'"`

	Flow     TaskFlow `json:"-" gorm:"type:text"`          // 执行流程
	CurrStep int      `json:"currStep" gorm:"default:'0'"` // 当前在执行的流程步骤

	// 任务每一步的执行超时(整个任务无超时控制)
	StepTimeout int `json:"stepTimeout" gorm:"default:'600';comment:'执行超时'"`

	// gorm 配置见 Migrate()
	Status string `json:"status" enums:"'pending','running','failed','complete','timeout'"` // 任务状态

	Message string `json:"message"` // 任务的状态描述信息，如失败原因等

	StartAt *time.Time `json:"startAt" gorm:"null;comment:'任务开始时间'"` // 任务开始时间
	EndAt   *time.Time `json:"endAt" gorm:"null;comment:'任务结束时间'"`   // 任务结束时间

	// 任务执行结果，如 add/change/delete 的资源数量、outputs 等
	Result TaskResult `json:"result" gorm:"type:json"` // 任务执行结果
}

func (Task) TableName() string {
	return "iac_task"
}

func (Task) DefaultTaskName() string {
	return ""
}

func (t *Task) Exited() bool {
	return t.IsExitedStatus(t.Status)
}

func (t *Task) Started() bool {
	return t.IsStartedStatus(t.Status)
}

func (Task) IsStartedStatus(status string) bool {
	// 注意：approving 状态的任务我们也认为其 started
	return !utils.InArrayStr([]string{TaskPending}, status)
}

func (Task) IsExitedStatus(status string) bool {
	return utils.InArrayStr([]string{TaskFailed, TaskComplete}, status)
}

func (t *Task) IsEffectTask() bool {
	return t.IsEffectTaskType(t.Type)
}

// IsEffectTaskType 是否产生实际数据变动的任务类型
func (Task) IsEffectTaskType(typ string) bool {
	return utils.StrInArray(typ, TaskTypeApply, TaskTypeDestroy)
}

func (Task) GetTaskNameByType(typ string) string {
	switch typ {
	case TaskTypePlan:
		return common.TaskTypePlanName
	case TaskTypeApply:
		return common.TaskTypeApplyName
	case TaskTypeDestroy:
		return common.TaskTypeDestroyName
	default:
		panic("invalid task type")
	}
}
func (t *Task) StateJsonPath() string {
	return path.Join(t.ProjectId.String(), t.EnvId.String(), t.Id.String(), runner.TFStateJsonFile)
}

func (t *Task) PlanJsonPath() string {
	return path.Join(t.ProjectId.String(), t.EnvId.String(), t.Id.String(), runner.TFPlanJsonFile)
}

func (t *Task) Migrate(sess *db.Session) (err error) {
	// 以下 column 通过 Migrate 来维护，确保新增加的 enum 生效
	columnDefines := []struct {
		column     string
		typeDefine string
	}{
		{
			"status",
			`ENUM('pending','running','approving','failed','complete','timeout') DEFAULT 'pending' COMMENT '作业状态'`,
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
	Name string   `json:"name,omitempty" yaml:"name" gorm:"size:32;not null"`
	Args StrSlice `json:"args,omitempty" yaml:"args" gorm:"type:text"`
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
	NextStep  Id         `json:"nextStep" gorm:"size:32;default:''"`
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

func (s *TaskStep) IsStarted() bool {
	return !utils.StrInArray(s.Status, TaskStepPending, TaskStepApproving)
}

func (s *TaskStep) IsExited() bool {
	return utils.StrInArray(s.Status, TaskStepRejected, TaskStepComplete, TaskStepFailed, TaskStepTimeout)
}

func (s *TaskStep) IsApproved() bool {
	if s.Status == TaskStepRejected {
		return false
	}
	// 只有 apply 和 destroy 步骤需要审批
	if utils.StrInArray(s.Type, TaskStepApply, TaskStepDestroy) && len(s.ApproverId) == 0 {
		return false
	}
	return true
}

func (s *TaskStep) GenLogPath() string {
	return path.Join(
		s.ProjectId.String(),
		s.EnvId.String(),
		s.TaskId.String(),
		fmt.Sprintf("step%d", s.Index),
		runner.TaskStepLogName,
	)
}
