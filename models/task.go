package models

import (
	"cloudiac/libs/db"
	"time"
)

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
	Guid         string     `json:"guid" gorm:"not null;comment:'任务guid'"`
	TaskName     string     `json:"taskName" gorm:"not null;comment:'任务名称'"`
	TemplateGuid string     `json:"templateGuid" gorm:"size:32;not null;comment:'模板GUID'"`
	TemplateId   uint       `json:"templateId" gorm:"size:32;not null;comment:'模板ID'"`
	TaskType     string     `json:"taskType" gorm:"type:enum('plan','apply');not null;comment:'作业类型'"`
	Status       string     `json:"status" gorm:"type:enum('pending','running','failed','complete','timeout');default:'pending';comment:'作业状态'"`
	BackendInfo  JSON       `json:"backendInfo" gorm:"type:json;null;comment:'执行信息'" json:"backend_info"`
	Timeout      int        `json:"timeout" gorm:"size:32;comment:'超时时长'"`
	Creator      uint       `json:"creator" gorm:"not null;comment:'创建人'"`
	StartAt      *time.Time `json:"startAt" gorm:"null;comment:'任务开始时间'"`
	EndAt        *time.Time `json:"endAt" gorm:"null;comment:'任务结束时间'"`
	CommitId     string     `json:"commitId" gorm:"null;comment:'COMMIT ID'"`
	CtServiceId  string     `json:"ctServiceId" form:"ctServiceId" comment:'runnerId'"`
}

func (Task) TableName() string {
	return "iac_task"
}

func (o Task) Migrate(sess *db.Session) (err error) {
	err = o.AddUniqueIndex(sess, "unique__guid", "guid")
	if err != nil {
		return err
	}
	return nil
}
