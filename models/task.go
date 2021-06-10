package models

import (
	"cloudiac/consts"
	"cloudiac/libs/db"
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
	Guid          string     `json:"guid" gorm:"not null;comment:'任务guid'"`
	Name          string     `json:"name" gorm:"not null;comment:'任务名称'"`
	TemplateGuid  string     `json:"templateGuid" gorm:"size:32;not null;comment:'模板GUID'"`
	TemplateId    uint       `json:"templateId" gorm:"size:32;not null;comment:'模板ID'"`
	TaskType      string     `json:"taskType"` // gorm 配置见 Migrate()
	Status        string     `json:"status"`   // gorm 配置见 Migrate()
	StatusDetail  string     `json:"statusDetail" gorm:"comment:'状态说明信息'"`
	Creator       uint       `json:"creator" gorm:"not null;comment:'创建人'"`
	StartAt       *time.Time `json:"startAt" gorm:"null;comment:'任务开始时间'"`
	EndAt         *time.Time `json:"endAt" gorm:"null;comment:'任务结束时间'"`
	CommitId      string     `json:"commitId" gorm:"null;comment:'COMMIT ID'"`
	CtServiceId   string     `json:"ctServiceId" gorm:"comment:'runnerId'"`
	Source        string     `json:"source" gorm:"null;comment:'来源(workflow等)'"`
	TransactionId string     `json:"transactionId" gorm:"null;comment:'流水号Id(workflow用)'"`
	Add           string     `json:"add" gorm:"default:0"`
	Change        string     `json:"change" gorm:"default:0"`
	Destroy       string     `json:"destroy" gorm:"default:0"`
	AllowApply    bool       `json:"allowApply" gorm:"default:false"`

	SourceVars  JSON            `json:"sourceVars" gorm:"type:json;null;comment:'来源参数(workflow等)'"`
	BackendInfo *TaskBackendInfo `json:"backendInfo" gorm:"type:json;null;comment:'执行信息'" json:"backend_info"`
}

func (Task) TableName() string {
	return "iac_task"
}

func (t *Task) Exited() bool {
	return utils.InArrayStr([]string{consts.TaskFailed, consts.TaskComplete, consts.TaskTimeout}, t.Status)
}

func (t *Task) Started() bool {
	return !utils.InArrayStr([]string{consts.TaskPending, consts.TaskAssigning}, t.Status)
}

func (t *Task) Migrate(sess *db.Session) (err error) {
	err = t.AddUniqueIndex(sess, "unique__guid", "guid")
	if err != nil {
		return err
	}

	// 以下 column 通过 Migrate 来维护，确保新增加的 enum 生效
	columnDefines := []struct {
		column     string
		typeDefine string
	}{
		{
			"status",
			`ENUM('pending','running','failed','complete','timeout','assigning') DEFAULT 'pending' COMMENT '作业状态'`,
		},
		{
			"task_type",
			`ENUM('plan','apply','destroy','pull','debug') NOT NULL COMMENT '作业类型'`,
		},
	}
	for _, cd := range columnDefines {
		if err := sess.DB().ModifyColumn(cd.column, cd.typeDefine).Error; err != nil {
			return err
		}
	}

	return nil
}
