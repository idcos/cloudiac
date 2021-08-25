package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func CreatePolicyGroup(tx *db.Session, group *models.PolicyGroup) (*models.PolicyGroup, e.Error) {
	if err := models.Create(tx, group); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.PolicyGroupAlreadyExist, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return group, nil
}

func GetPolicyGroupById(tx *db.Session, id models.Id) (*models.PolicyGroup, e.Error) {
	group := models.PolicyGroup{}
	if err := tx.Model(models.PolicyGroup{}).Where("id = ?", id).First(&group); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.PolicyGroupNotExist, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &group, nil
}
