// Copyright 2021 CloudJ Company Limited. All rights reserved.

package models

import (
	"cloudiac/common"
	"cloudiac/portal/libs/db"
	"cloudiac/runner"
	"cloudiac/utils"
	"database/sql/driver"
	"path"
)

type TaskVariables []VariableBody

func (v TaskVariables) Value() (driver.Value, error) {
	return MarshalValue(v)
}

func (v *TaskVariables) Scan(value interface{}) error {
	return UnmarshalValue(value, v)
}

type TaskResult struct {
	ResAdded     *int `json:"resAdded"` // 该值为 nil 表示无资源变更数据(区别于 0)
	ResChanged   *int `json:"resChanged"`
	ResDestroyed *int `json:"resDestroyed"`

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
	TaskTypeScan    = common.TaskTypeScan
	TaskTypeParse   = common.TaskTypeParse

	TaskPending   = common.TaskPending
	TaskRunning   = common.TaskRunning
	TaskApproving = common.TaskApproving
	TaskRejected  = common.TaskRejected
	TaskFailed    = common.TaskFailed
	TaskComplete  = common.TaskComplete
)

type Tasker interface {
	GetId() Id
	GetRunnerId() string
	GetStepTimeout() int
	Exited() bool
	Started() bool
	IsStartedStatus(status string) bool
	IsExitedStatus(status string) bool
	IsEffectTask() bool
	IsEffectTaskType(typ string) bool
	GetTaskNameByType(typ string) string
}

// Task 部署任务
type Task struct {
	BaseTask

	OrgId     Id `json:"orgId" gorm:"size:32;not null"`     // 组织ID
	ProjectId Id `json:"projectId" gorm:"size:32;not null"` // 项目ID
	TplId     Id `json:"tplId" gorm:"size:32;not null"`     // 模板ID
	EnvId     Id `json:"envId" gorm:"size:32;not null"`     // 环境ID

	Name      string `json:"name" gorm:"not null;comment:任务名称"` // 任务名称
	CreatorId Id     `json:"creatorId" gorm:"size:32;not null"` // 创建人ID

	RepoAddr string `json:"repoAddr" gorm:"not null"`
	Revision string `json:"revision" gorm:"not null"`
	CommitId string `json:"commitId" gorm:"not null"` // 创建任务时 revision 对应的 commit id

	Workdir      string   `json:"workdir" gorm:"default:''"`
	Playbook     string   `json:"playbook" gorm:"default:''"`
	TfVarsFile   string   `json:"tfVarsFile" gorm:"default:''"`
	TfVersion    string   `json:"tfVersion" gorm:"default:''"`
	PlayVarsFile string   `json:"playVarsFile" gorm:"default:''"`
	Targets      StrSlice `json:"targets" gorm:"type:json"` // 指定 terraform target 参数

	Variables TaskVariables `json:"variables" gorm:"type:json"` // 本次执行使用的所有变量(继承、覆盖计算之后的)

	StatePath string `json:"statePath" gorm:"not null"`

	// 扩展属性，包括 source, transitionId 等
	ExtraData JSON      `json:"extraData" gorm:"type:json"` // 扩展字段，用于存储外部服务调用时的信息

	KeyId           Id   `json:"keyId" gorm:"size32"` // 部署密钥ID
	AutoApprove     bool `json:"autoApproval" gorm:"default:false"`
	StopOnViolation bool `json:"stopOnViolation" gorm:"default:false"`

	// 任务执行结果，如 add/change/delete 的资源数量、outputs 等
	Result TaskResult `json:"result" gorm:"type:json"` // 任务执行结果

	RetryNumber int  `json:"retryNumber" gorm:"size:32;default:0"` // 任务重试次数
	RetryDelay  int  `json:"retryDelay" gorm:"size:32;default:0"`  // 每次任务重试时间，单位为秒
	RetryAble   bool `json:"retryAble" gorm:"default:false"`
}

func (Task) TableName() string {
	return "iac_task"
}

func (Task) DefaultTaskName() string {
	return ""
}

func (t *BaseTask) GetId() Id {
	return t.Id
}

func (t *BaseTask) GetRunnerId() string {
	return t.RunnerId
}

func (t *BaseTask) GetStepTimeout() int {
	return t.StepTimeout
}

func (t *BaseTask) Exited() bool {
	return t.IsExitedStatus(t.Status)
}

func (t *BaseTask) Started() bool {
	return t.IsStartedStatus(t.Status)
}

func (BaseTask) IsStartedStatus(status string) bool {
	// 注意：approving 状态的任务我们也认为其 started
	return !utils.InArrayStr([]string{TaskPending}, status)
}

func (BaseTask) IsExitedStatus(status string) bool {
	return utils.InArrayStr([]string{TaskFailed, TaskRejected, TaskComplete}, status)
}

func (t *BaseTask) IsEffectTask() bool {
	return t.IsEffectTaskType(t.Type)
}

// IsEffectTaskType 是否产生实际数据变动的任务类型
func (BaseTask) IsEffectTaskType(typ string) bool {
	return utils.StrInArray(typ, TaskTypeApply, TaskTypeDestroy)
}

func (BaseTask) GetTaskNameByType(typ string) string {
	switch typ {
	case TaskTypePlan:
		return common.TaskTypePlanName
	case TaskTypeApply:
		return common.TaskTypeApplyName
	case TaskTypeDestroy:
		return common.TaskTypeDestroyName
	case TaskTypeScan:
		return common.TaskTypeScanName
	case TaskTypeParse:
		return common.TaskTypeParse
	default:
		panic("invalid task type")
	}
}
func (t *Task) StateJsonPath() string {
	return path.Join(t.ProjectId.String(), t.EnvId.String(), t.Id.String(), runner.TFStateJsonFile)
}

func (t *Task) ProviderSchemaJsonPath() string {
	return path.Join(t.ProjectId.String(), t.EnvId.String(), t.Id.String(), runner.TFProviderSchema)
}

func (t *Task) PlanJsonPath() string {
	return path.Join(t.ProjectId.String(), t.EnvId.String(), t.Id.String(), runner.TFPlanJsonFile)
}

func (t *Task) TfParseJsonPath() string {
	return path.Join(t.ProjectId.String(), t.EnvId.String(), t.Id.String(), runner.TerrascanJsonFile)
}

func (t *Task) TfResultJsonPath() string {
	return path.Join(t.ProjectId.String(), t.EnvId.String(), t.Id.String(), runner.TerrascanResultFile)
}

func (t *Task) HideSensitiveVariable() {
	for index, v := range t.Variables {
		if v.Sensitive {
			t.Variables[index].Value = ""
		}
	}
}

func (t *Task) Migrate(sess *db.Session) (err error) {
	if err := sess.ModifyModelColumn(t, "status"); err != nil {
		return err
	}
	return nil
}
