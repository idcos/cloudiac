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
	Name         string     `json:"name" gorm:"not null;comment:'任务名称'"`
	TemplateGuid string     `json:"templateGuid" gorm:"size:32;not null;comment:'模板GUID'"`
	TemplateId   uint       `json:"templateId" gorm:"size:32;not null;comment:'模板ID'"`
	TaskType     string     `json:"taskType" gorm:"type:enum('plan','apply');not null;comment:'作业类型'"`
	Status       string     `json:"status" gorm:"type:enum('pending','running','failed','complete','timeout');default:'pending';comment:'作业状态'"`
	BackendInfo  JSON       `json:"backendInfo" gorm:"type:json;null;comment:'执行信息'" json:"backend_info"`
	Creator      uint       `json:"creator" gorm:"not null;comment:'创建人'"`
	StartAt      *time.Time `json:"startAt" gorm:"null;comment:'任务开始时间'"`
	EndAt        *time.Time `json:"endAt" gorm:"null;comment:'任务结束时间'"`
	CommitId     string     `json:"commitId" gorm:"null;comment:'COMMIT ID'"`
	CtServiceId  string     `json:"ctServiceId" gorm:"comment:'runnerId'"`
	Source       string     `json:"source" gorm:"null;comment:'来源(workflow等)'"`
	SourceVars   JSON       `json:"sourceVars" gorm:"type:json;null;comment:'来源参数(workflow等)'"`
	TransactionId string `json:"transactionId" gorm:"null;comment:'流水号Id(workflow用)'"`
	Add          string   	`json:"add" gorm:"default:0"`
	Change       string		`json:"change" gorm:"default:0"`
	Destroy 	 string		`json:"destroy" gorm:"default:0"`
	AllowApply   bool		`json:"allowApply" gorm:"default:false"`
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
