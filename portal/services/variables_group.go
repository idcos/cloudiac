package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

func CreateVariableGroup(tx *db.Session, group models.VariableGroup) (models.VariableGroup, e.Error) {
	if group.Id == "" {
		group.Id = models.NewId("vg")
	}
	if err := models.Create(tx, &group); err != nil {
		if e.IsDuplicate(err) {
			return group, e.New(e.VariableGroupAlreadyExist, err)
		}
		return group, e.AutoNew(err, e.DBError)
	}
	return group, nil
}

func SearchVariableGroup(dbSess *db.Session, orgId models.Id, q string) *db.Session {
	query := dbSess.Model(models.VariableGroup{}).Where("org_id = ?", orgId)
	if q != "" {
		query = query.WhereLike("name", q)
	}
	return query
}

func UpdateVariableGroup(tx *db.Session, id models.Id, attrs models.Attrs) e.Error {
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.VariableGroup{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return e.New(e.VariableGroupAliasDuplicate)
		} else if e.IsRecordNotFound(err) {
			return e.New(e.VariableGroupNotExist)
		}
		return e.New(e.DBError, fmt.Errorf("update variable group error: %v", err))
	}
	return nil
}


func DeleteVariableGroup(tx *db.Session, vgId models.Id) e.Error {
	//删除变量组
	if _, err := tx.Where("id = ?", vgId).Delete(&models.VariableGroup{}); err != nil {
		return e.New(e.DBError, err)
	}

	//删除变量组与实例之间的关系
	if _, err := tx.Where("var_group_id = ?", vgId).Delete(&models.VariableGroupRel{}); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}


func DetailVariableGroup(dbSess *db.Session, vgId, orgId models.Id) *db.Session {
	return dbSess.Model(&models.VariableGroup{}).
		Where("id = ?", vgId).
		Where("org_id = ?", orgId)
}

