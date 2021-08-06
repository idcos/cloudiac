// Copyright 2021 CloudJ Company Limited. All rights reserved.

package models

import (
	"cloudiac/portal/libs/db"
)

type OperationLog struct {
	BaseModel

	UserID        Id     `json:"userId" form:"userId" gorm:"size:32"`
	Username      string `json:"username" form:"username" `
	UserAddr      string `json:"userAddr" form:"userAddr" `
	OperationAt   Time   `json:"operationAt"  gorm:"type:datetime" form:"operationAt" `
	OperationType string `json:"operationType" form:"operationType" `
	OperationInfo string `json:"operationInfo" form:"operationInfo" `
	Desc          JSON   `json:"desc" form:"desc" gorm:"type:text"`
}

func (o *OperationLog) InsertLog() error {
	return db.Get().Insert(o)
}

func (OperationLog) TableName() string {
	return "iac_operation_log"
}
