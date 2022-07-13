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
	query := dbSess.Model(&models.Policy{}).Joins(`join iac_policy_group on iac_policy_group.id = iac_policy.group_id`)

	query = query.Where("iac_policy_group.enabled = ?", 1)
	if len(orgIds) > 0 {
		query = query.Where(`iac_policy.org_id IN (?)`, orgIds)
	}

	return query.Count()
}

func GetPGStackEnabledCount(dbSess *db.Session, orgIds []string) (int64, error) {
	subQuery := dbSess.Model(&models.PolicyRel{}).Select("DISTINCT(tpl_id) as id ")
	subQuery = subQuery.Joins(`join iac_policy_group on iac_policy_group.id = iac_policy_rel.group_id`)
	subQuery = subQuery.Where(`iac_policy_group.enabled = ?`, 1)

	query := dbSess.Model(&models.Template{}).Where(`id IN (?)`, subQuery.Expr())
	query = query.Where("iac_template.status = ?", models.Enable)
	query = query.Where("iac_template.policy_enable = ?", 1)
	if len(orgIds) > 0 {
		query = query.Where(`iac_template.org_id IN (?)`, orgIds)
	}

	return query.Count()
}

func GetPGEnvEnabledCount(dbSess *db.Session, orgIds []string) (int64, error) {
	subQuery := dbSess.Model(&models.PolicyRel{}).Select("DISTINCT(env_id) as id ")
	subQuery = subQuery.Joins(`join iac_policy_group on iac_policy_group.id = iac_policy_rel.group_id`)
	subQuery = subQuery.Where(`iac_policy_group.enabled = ?`, 1)

	query := dbSess.Model(&models.Env{}).Where(`id IN (?)`, subQuery.Expr())
	query = query.Where(`(status = ? OR status = ? OR task_status = ? OR task_status = ?)`, models.EnvStatusActive, models.EnvStatusFailed, models.TaskApproving, models.TaskRunning)
	query = query.Where("iac_env.archived = ?", 0)
	query = query.Where("iac_env.policy_enable = ?", 1)
	if len(orgIds) > 0 {
		query = query.Where(`iac_env.org_id IN (?)`, orgIds)
	}
	query = query.Where("iac_env.updated_at > DATE_SUB(CURDATE(), INTERVAL ? DAY)", 7)

	return query.Count()
}

func GetPGStackNGCount(dbSess *db.Session, orgIds []string) (int64, error) {
	subQuery := dbSess.Model(&models.PolicyResult{}).Select("tpl_id, MAX(start_at)")
	subQuery = subQuery.Where(`iac_policy_result.status = ?`, models.EnvStatusFailed)
	subQuery = subQuery.Group("tpl_id")

	query := dbSess.Model(&models.Template{}).Joins("inner join (?) as t on t.tpl_id = iac_template.id", subQuery.Expr())
	query = query.Where("iac_template.status = ?", models.Enable)
	query = query.Where("iac_template.policy_enable = ?", 1)
	if len(orgIds) > 0 {
		query = query.Where(`iac_template.org_id IN (?)`, orgIds)
	}

	return query.Count()
}

func GetPGEnvNGCount(dbSess *db.Session, orgIds []string) (int64, error) {
	subQuery := dbSess.Model(&models.PolicyResult{}).Select("env_id, MAX(start_at)")
	subQuery = subQuery.Where(`iac_policy_result.status = ?`, models.EnvStatusFailed)
	subQuery = subQuery.Group("env_id")

	query := dbSess.Model(&models.Env{}).Joins("inner join (?) as t on t.env_id = iac_env.id", subQuery.Expr())
	query = query.Where(`(status = ? OR status = ? OR task_status = ? OR task_status = ?)`, models.EnvStatusActive, models.EnvStatusFailed, models.TaskApproving, models.TaskRunning)
	query = query.Where("iac_env.archived = ?", 0)
	query = query.Where("iac_env.policy_enable = ?", 1)
	if len(orgIds) > 0 {
		query = query.Where(`iac_env.org_id IN (?)`, orgIds)
	}
	query = query.Where("iac_env.updated_at > DATE_SUB(CURDATE(), INTERVAL ? DAY)", 7)

	return query.Count()
}
