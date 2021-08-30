package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
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

func SearchPolicyGroup(dbSess *db.Session, orgId models.Id, q string) *db.Session {
	pgTable := models.PolicyGroup{}.TableName()
	query := dbSess.Table(pgTable).
		Joins(fmt.Sprintf("left join (%s) as p on p.group_id = %s.id",
			fmt.Sprintf("select count(1) as policy_count,group_id from %s",
				models.Policy{}.TableName()), pgTable))
		//Where(fmt.Sprintf("%s.org_id = ?", pgTable), orgId)
	if q != "" {
		qs := "%" + q + "%"
		query = query.Where(fmt.Sprintf("%s.name like ?", pgTable), qs)
	}
	return query.LazySelectAppend(fmt.Sprintf("%s.*,p.policy_count", pgTable))
}

func UpdatePolicyGroup(query *db.Session, group *models.PolicyGroup, attr models.Attrs) e.Error {
	if _, err := models.UpdateAttr(query, group, attr); err != nil {
		if e.IsDuplicate(err) {
			return e.New(e.PolicyGroupAlreadyExist, err)
		}
		return e.New(e.DBError, err)
	}
	return nil
}

func DeletePolicyGroup(tx *db.Session, groupId models.Id) e.Error {
	if _, err := tx.Where("id = ?", groupId).
		Delete(&models.PolicyGroup{}); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

func DetailPolicyGroup(dbSess *db.Session, groupId models.Id) (*models.PolicyGroup, e.Error) {
	pg := &models.PolicyGroup{}
	if err := dbSess.
		Where("id = ?", groupId).
		First(pg); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.PolicyGroupAlreadyExist, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return pg, nil
}
