// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/utils/logs"
)

func InsertUserOperateLog(operatorId, orgId, objectId models.Id, objectType, action string, attr models.ResAttrs) {
	session := db.Get()
	err := models.Create(session, &models.UserOperationLog{
		ObjectType: objectType,
		ObjectId:   objectId,
		Action:     action,
		OperatorId: operatorId,
		OrgId:      orgId,
		Attribute:  attr,
	})
	if err != nil {
		logs.Get().Errorf("operate log insert err: %v", err)
	}
}
