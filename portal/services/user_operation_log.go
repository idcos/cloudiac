package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func InsertUserOperateLog(operatorId, orgId, objectId models.Id, objectType, action string, attr models.ResAttrs) error {
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
		return err
	}

	return nil
}
