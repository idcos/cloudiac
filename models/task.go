package models

import "cloudiac/libs/db"

const (
	PENDING  = "pending"
	RUNNING  = "running"
	FAILED   = "failed"
	COMPLETE = "complete"
	TIMEOUT  = "timeout"
)

const (
	PLAN  = "plan"
	APPLY = "apply"
)

type Task struct {
	SoftDeleteModel
	Guid     string `json:"guid" form:"guid" `
	TaskName     string `json:"taskName" form:"taskName" gorm:"not null;comment:'任务名称'"`
	TemplateGuid string `gorm:"size:32;not null;comment:'模板GUID'"`
	TemplateId   uint64 `gorm:"size:32;not null;comment:'模板ID'"`
	TaskType     string `gorm:"type:enum('plan','apply');not null;comment:'作业类型'"`
	Status       string `gorm:"type:enum('pending','running','failed','complete','timeout');default:'pending';comment:'作业状态'"`
	BackendInfo  JSON   `gorm:"type:json;null;comment:'执行信息'" json:"backend_info"`
	Timeout      int    `json:"timeout" gorm:"size:32;comment:'超时时长'"`
	Creator      uint   `json:"creator" grom:"not null;comment:'创建人'"`
}

func (Task) TableName() string {
	return "iac_task"
}

func (o Task) Migrate(sess *db.Session) (err error) {
	return nil
}
