package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func CreatePolicyRel(tx *db.Session, group *models.PolicyRel) (*models.PolicyRel, e.Error) {
	if err := models.Create(tx, group); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.PolicyGroupAlreadyExist, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return group, nil
}
