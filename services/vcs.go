package services

import (
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"
	"fmt"
)

func CreateVcs(tx *db.Session, vcs models.Vcs) (*models.Vcs, e.Error) {
	if err := models.Create(tx, &vcs); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return &vcs, nil
}

func UpdateVcs(tx *db.Session, id uint, attrs models.Attrs) (vcs *models.Vcs, er e.Error) {
	vcs = &models.Vcs{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Vcs{}, attrs); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("update vcs error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(vcs); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query vcs error: %v", err))
	}
	return
}

func QueryVcs(orgId uint, query *db.Session) *db.Session {
	return query.Model(&models.Vcs{}).Where("org_id = ?", orgId)
}

func DeleteVcs(tx *db.Session, id uint) e.Error {
	if _, err := tx.Where("id = ?", id).Delete(&models.Vcs{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete vcs error: %v", err))
	}
	return nil
}
