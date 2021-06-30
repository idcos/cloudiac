package models

import (
	"cloudiac/portal/libs/db"
	"time"
)

type OperationLog struct {
	BaseModel

	UserID        uint      `json:"userId" form:"userId" `
	Username      string    `json:"username" form:"username" `
	UserAddr      string    `json:"userAddr" form:"userAddr" `
	OperationAt   time.Time `json:"operationAt" form:"operationAt" `
	OperationType string    `json:"operationType" form:"operationType" `
	OperationInfo string    `json:"operationInfo" form:"operationInfo" `
	Desc          JSON      `json:"desc" form:"desc" `
}

func (o *OperationLog) InsertLog() error {
	return db.Get().Insert(o)
}

func (OperationLog) TableName() string {
	return "iac_operation_log"
}
