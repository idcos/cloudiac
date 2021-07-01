package services

import (
	"fmt"
	//"errors"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func CreateResourceAccount(tx *db.Session, rsAccount *models.ResourceAccount) (*models.ResourceAccount, e.Error) {
	if err := models.Create(tx, rsAccount); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.NameDuplicate, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return rsAccount, nil
}

func UpdateResourceAccount(tx *db.Session, id models.Id, attrs models.Attrs) (rsAccount *models.ResourceAccount, re e.Error) {
	rsAccount = &models.ResourceAccount{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.ResourceAccount{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.NameDuplicate, err)
		}
		return nil, e.New(e.DBError, fmt.Errorf("update resourceAccount error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(rsAccount); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query resourceAccount error: %v", err))
	}
	return
}

func DeleteResourceAccount(tx *db.Session, id models.Id, orgId models.Id) e.Error {
	if _, err := tx.Where("id = ? AND org_id = ?", id, orgId).Delete(&models.ResourceAccount{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete resourceAccount error: %v", err))
	}
	return nil
}

func GetResourceAccountById(tx *db.Session, id models.Id) (*models.ResourceAccount, e.Error) {
	r := models.ResourceAccount{}
	if err := tx.Where("id = ?", id).First(&r); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.ObjectNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &r, nil
}

func GetResourceAccountByName(tx *db.Session, name string) (*models.ResourceAccount, error) {
	r := models.ResourceAccount{}
	if err := tx.Where("name = ?", name).First(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func FindResourceAccount(query *db.Session) (rsAccount []*models.ResourceAccount, err error) {
	err = query.Find(&rsAccount)
	return
}

func QueryResourceAccount(query *db.Session) *db.Session {
	return query.Model(&models.ResourceAccount{})
}

func CreateCtResourceMap(tx *db.Session, ctResourceMap models.CtResourceMap) (*models.CtResourceMap, e.Error) {
	if err := models.Create(tx, &ctResourceMap); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.NameDuplicate, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &ctResourceMap, nil
}

func FindCtResourceMap(query *db.Session, rsAccountId models.Id) (ctServiceIds []string, err error) {
	ctResourceMap := []*models.CtResourceMap{}
	if err := query.Where("resource_account_id = ?", rsAccountId).Find(&ctResourceMap); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	for _, r := range ctResourceMap {
		ctServiceIds = append(ctServiceIds, r.CtServiceId)
	}
	return
}

func DeleteCtResourceMap(tx *db.Session, rsAccountId models.Id) e.Error {
	if _, err := tx.Where("resource_account_id = ?", rsAccountId).Delete(&models.CtResourceMap{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete CtResourceMap error: %v", err))
	}
	return nil
}
