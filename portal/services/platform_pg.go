package services

import (
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func GetPolicyGroupCount(dbSess *db.Session, orgIds []string) (int64, error) {
	query := dbSess.Model(&models.PolicyGroup{}).Where("enabled = ?", 1)
	if len(orgIds) > 0 {
		query = query.Where(`org_id IN (?)`, orgIds)
	}

	return query.Count()
}

func GetPolicyCount(dbSess *db.Session, orgIds []string) (int64, error) {
	query := dbSess.Model(&models.Policy{}).Joins(`iac_policy_group on iac_policy_group.id = iac_policy.group_id`)

	query = query.Where("iac_policy_group.enabled = ?", 1)
	if len(orgIds) > 0 {
		query = query.Where(`iac_policy.org_id IN (?)`, orgIds)
	}

	return query.Count()
}
