// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package models

import (
	"cloudiac/portal/libs/db"
)

type OperationLog struct {
	BaseModel

	UserID          Id     `json:"userId" form:"userId" gorm:"size:32"`
	Username        string `json:"username" form:"username" `
	UserAddr        string `json:"userAddr" form:"userAddr" `
	OperationAt     Time   `json:"operationAt"  gorm:"type:datetime" form:"operationAt" `
	OperationUrl    Text   `json:"operationUrl" gorm:"type:text" form:"operationUrl" `
	OperationType   string `json:"operationType" form:"operationType" `
	OperationInfo   string `json:"operationInfo" form:"operationInfo" `
	OperationStatus int    `json:"operationStatus" form:"operationStatus" `
	Desc            JSON   `json:"desc" form:"desc" gorm:"type:text"`
}

func (o *OperationLog) InsertLog() error {
	return db.Get().Insert(o)
}

func (OperationLog) TableName() string {
	return "iac_operation_log"
}

type UserOperationLog struct {
	TimedModel
	ObjectType string   `json:"objectType"`
	ObjectId   Id       `json:"objectId"`
	ObjectName string   `json:"objectName"`
	Action     string   `json:"action"`
	OperatorId Id       `json:"operatorId" gorm:"size:32"`
	OrgId      Id       `json:"orgId" gorm:"size:32"`
	Attribute  ResAttrs `json:"attribute" gorm:"type:text"`
}

func (UserOperationLog) TableName() string {
	return "iac_user_operation_log"
}
