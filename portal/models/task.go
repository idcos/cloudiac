package models

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/db"
	"cloudiac/utils"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type TaskBackendInfo struct {
	BackendUrl  string `json:"backend_url"`
	CtServiceId string `json:"ct_service_id"`
	LogFile     string `json:"log_file"`
	ContainerId string `json:"container_id"`
}

func (b *TaskBackendInfo) Value() (driver.Value, error) {
	if b == nil {
		return nil, nil
	}
	bs, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (b *TaskBackendInfo) Scan(value interface{}) error {
	if value == nil {
		*b = TaskBackendInfo{}
		return nil
	}

	bs, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type %T, value: %T", value, value)
	}
	return json.Unmarshal(bs, b)
}

const (
	PENDING   = consts.TaskPending
	RUNNING   = consts.TaskRunning
	ASSIGNING = consts.TaskAssigning
	FAILED    = consts.TaskFailed
	COMPLETE  = consts.TaskComplete
	TIMEOUT   = consts.TaskTimeout
)

var TaskStatusList = []string{PENDING, RUNNING, ASSIGNING, FAILED, COMPLETE, TIMEOUT}

const (
	PLAN  = consts.TaskPlan
	APPLY = consts.TaskApply
)

type Task struct {
	SoftDeleteModel

	OrgId     Id `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id `json:"projectId" gorm:"size:32;not null"`
	EnvId     Id `json:"envId" gorm:"size:32;not null"`

	Name      string `json:"name" gorm:"not null;comment:'任务名称'"`
	CreatorId Id   `json:"size:32;creatorId"`
	RunnerId  string `json:"runnerId" gorm:"not null"`
	CommitId  string `json:"commitId" gorm:"not null"`
	Status    string `json:"status"`  // gorm 配置见 Migrate()
	Message   string `json:"message"` // 任务的状态描述信息，如失败原因

	Flow     string `json:"-" gorm:"type:text"`
	CurrStep int    `json:"currStep" gorm:"default:'0'"` // 当前在执行的流程步骤

	StartAt *time.Time `json:"startAt" gorm:"null;comment:'任务开始时间'"`
	EndAt   *time.Time `json:"endAt" gorm:"null;comment:'任务结束时间'"`

	// TODO JSON 类型改为具体结构体

	// 本地执行使用的所有变量(继承、覆盖计算之后的)
	Variables JSON `json:"variables" gorm:"type:json"`

	// 任务执行结果: add/change/delete 资源数量
	Result JSON `json:"result"`

	// 扩展属性，包括 source, transitionId 等
	Extra JSON `json:"extra" gorm:"type:json"`
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
	return !utils.InArrayStr([]string{consts.TaskPending, consts.TaskAssigning}, status)
}

func (Task) IsExitedStatus(status string) bool {
	return utils.InArrayStr([]string{consts.TaskFailed, consts.TaskComplete, consts.TaskTimeout}, status)
}

func (t *Task) Migrate(sess *db.Session) (err error) {
	// 以下 column 通过 Migrate 来维护，确保新增加的 enum 生效
	columnDefines := []struct {
		column     string
		typeDefine string
	}{
		{
			"status",
			`ENUM('pending','running','failed','complete','timeout','assigning') DEFAULT 'pending' COMMENT '作业状态'`,
		},
	}
	for _, cd := range columnDefines {
		if err := sess.DB().ModifyColumn(cd.column, cd.typeDefine).Error; err != nil {
			return err
		}
	}

	return nil
}

type TaskStep struct {
	BaseModel
	OrgId     Id   `json:"orgId" gorm:"size:32;not null"`
	ProjectId Id   `json:"projectId" gorm:"size:32;not null"`
	TaskId    Id   `json:"taskId" gorm:"size:32;not null"`
	Index     Id   `json:"index" gorm:"size:32;not null"`
	Type      string `json:"type" gorm:"size:16"`
	Status    string `json:"status" gorm:"type:enum('pending','running','failed','done')"`
	LogPath   string `json:"logPath" gorm:""`
}

func (TaskStep) TableName() string {
	return "iac_task_step"
}
