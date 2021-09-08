// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/runner"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func CreatePolicy(tx *db.Session, policy *models.Policy) (*models.Policy, e.Error) {
	if err := models.Create(tx, policy); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.PolicyAlreadyExist, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return policy, nil
}

// GetPolicyReferenceId 生成策略ID
// reference id = "iac" + policy type + creator scope + max id
func GetPolicyReferenceId(query *db.Session, policy *models.Policy) (string, e.Error) {
	typ := "iac"
	if policy.PolicyType != "" {
		typ = policy.PolicyType
	}
	lastId := 0
	// query max id by type
	po := models.Policy{}
	if err := query.Model(models.Policy{}).Where("reference_id LIKE ?", "iac_"+typ+"%").
		Order("length(reference_id) DESC, reference_id DESC").Last(&po); err != nil && !e.IsRecordNotFound(err) {
		return "", e.AutoNew(err, e.DBError)
	}
	idx := strings.LastIndex(po.ReferenceId, "_")
	if idx != -1 {
		lastId, _ = strconv.Atoi(po.ReferenceId[idx+1:])
	}

	// internal or public
	scope := "public"

	return fmt.Sprintf("%s_%s_%s_%d", "iac", typ, scope, lastId+1), nil
}

func GetPolicyById(tx *db.Session, id models.Id) (*models.Policy, e.Error) {
	po := models.Policy{}
	if err := tx.Model(models.Policy{}).Where("id = ?", id).First(&po); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.PolicyNotExist, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &po, nil
}

func GetTaskPolicies(query *db.Session, taskId models.Id) ([]runner.TaskPolicy, e.Error) {
	var taskPolicies []runner.TaskPolicy
	policies, err := GetValidTaskPolicyIds(query, taskId)
	if err != nil {
		return nil, err
	}
	for _, policyId := range policies {
		policy, err := GetPolicyById(query, policyId)
		if err != nil {
			return nil, err
		}
		category := "general"
		group, _ := GetPolicyGroupById(query, policy.GroupId)
		if group != nil {
			category = group.Name
		}
		meta := map[string]interface{}{
			"name":          policy.RuleName,
			"file":          "policy.rego",
			"policy_type":   policy.PolicyType,
			"resource_type": policy.ResourceType,
			"severity":      strings.ToUpper(policy.Severity),
			"reference_id":  policy.ReferenceId,
			"category":      category,
			"version":       policy.Revision,
			"id":            string(policy.Id),
		}
		taskPolicies = append(taskPolicies, runner.TaskPolicy{
			PolicyId: string(policyId),
			Meta:     meta,
			Rego:     policy.Rego,
		})
	}
	return taskPolicies, nil
}

// GetValidTaskPolicyIds 获取策略关联的策略ID列表
func GetValidTaskPolicyIds(query *db.Session, taskId models.Id) ([]models.Id, e.Error) {
	var (
		policies  []models.Policy
		policyIds []models.Id
		envId     models.Id
		tplId     models.Id
		err       e.Error
	)
	if task, err := GetTask(query, taskId); err != nil {
		if e.IsRecordNotFound(err) {
			if scantask, err := GetScanTaskById(query, taskId); err != nil {
				return nil, err
			} else {
				envId = scantask.EnvId
				tplId = scantask.TplId
			}
		} else {
			return nil, err
		}
	} else {
		envId = task.EnvId
		tplId = task.TplId
	}

	if envId != "" {
		policies, err = GetPoliciesByEnvId(query, envId)
		if err != nil {
			return nil, err
		}
	} else {
		policies, err = GetPoliciesByTemplateId(query, tplId)
		if err != nil {
			return nil, err
		}
	}

	for _, policy := range policies {
		policyIds = append(policyIds, policy.Id)
	}

	return policyIds, nil
}

func suppressedQuery(query *db.Session) *db.Session {
	return query.Select("iac_policy.id").Table(models.PolicyRel{}.TableName()).
		Joins("join iac_policy on iac_policy.group_id = iac_policy_rel.group_id").
		Joins("join iac_policy_group on iac_policy_group.id = iac_policy_rel.group_id").
		Where("iac_policy.id not in (select policy_id from iac_policy_suppress WHERE env_id != '')").
		Where("iac_policy.id not in (select policy_id from iac_policy_suppress WHERE tpl_id != '')").
		Where("iac_policy.enabled = 1 and iac_policy_group.enabled = 1")
}

func GetPoliciesByEnvId(query *db.Session, envId models.Id) ([]models.Policy, e.Error) {
	subQuery := suppressedQuery(query).Where("iac_policy_rel.env_id = ?", envId)
	var policies []models.Policy
	if err := query.Model(models.Policy{}).Where("iac_policy.id in (?)", subQuery.Expr()).Find(&policies); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.PolicyNotExist, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return policies, nil
}

func GetPoliciesByTemplateId(query *db.Session, tplId models.Id) ([]models.Policy, e.Error) {
	subQuery := suppressedQuery(query).Where("iac_policy_rel.tpl_id = ?", tplId)
	var policies []models.Policy
	if err := query.Model(models.Policy{}).Where("iac_policy.id in (?)", subQuery.Expr()).Find(&policies); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.PolicyNotExist, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return policies, nil
}

func UpdatePolicy(tx *db.Session, policy *models.Policy, attr models.Attrs) (int64, e.Error) {
	affected, err := models.UpdateAttr(tx, policy, attr)
	if err != nil {
		if e.IsDuplicate(err) {
			return affected, e.New(e.PolicyGroupAlreadyExist, err)
		}
		return affected, e.New(e.DBError, err)
	}
	return affected, nil
}

//RemovePoliciesGroupRelation 移除策略组和策略的关系
func RemovePoliciesGroupRelation(tx *db.Session, groupId models.Id) e.Error {
	if _, err := UpdatePolicy(tx.Where("group_id = ?", groupId),
		&models.Policy{}, models.Attrs{"group_id": ""}); err != nil {
		return err
	}
	return nil
}

func SearchPolicy(dbSess *db.Session, form *forms.SearchPolicyForm) *db.Session {
	pTable := models.Policy{}.TableName()
	query := dbSess.Table(pTable)
	if len(form.GroupId) > 0 {
		query = query.Where(fmt.Sprintf("%s.group_id in (?)", pTable), form.GroupId)
	}

	if form.Severity != "" {
		query = query.Where(fmt.Sprintf("%s.severity = ?", pTable), form.Severity)
	}

	if form.Q != "" {
		qs := "%" + form.Q + "%"
		query = query.Where(fmt.Sprintf("%s.name like ?", pTable), qs)
	}

	query = query.Joins("left join iac_policy_group as g on g.id = iac_policy.group_id").
		LazySelectAppend("iac_policy.*,g.name as group_name")

	query = query.Joins("left join iac_user as u on u.id = iac_policy.creator_id").
		LazySelectAppend("iac_policy.*,u.name as creator")

	return query
}

func DeletePolicy(dbSess *db.Session, id models.Id) (interface{}, e.Error) {
	if _, err := dbSess.
		Where("id = ?", id).
		Delete(&models.Policy{}); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return nil, nil
}

func DetailPolicy(dbSess *db.Session, id models.Id) (interface{}, e.Error) {
	p := models.Policy{}
	if err := dbSess.Table(models.Policy{}.TableName()).
		Where("id = ?", id).
		First(&p); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.DBError, fmt.Errorf("polict not found id: %s", id))
		}
		return nil, e.New(e.DBError, err)
	}
	return p, nil
}

func SearchPolicySuppress(query *db.Session, id models.Id) *db.Session {
	t := models.PolicyRel{}.TableName()
	q := query.Model(models.PolicyRel{}.TableName()).
		Select(fmt.Sprintf("DISTINCT iac_policy.id as dist,%s.*,iac_policy.id as policy_id,if(%s.env_id='',%s.tpl_id,%s.env_id)as target_id,if(%s.env_id='','template','env')as target_type,if(%s.env_id = '',iac_template.name,iac_env.name) as target_name,if(e.id is null, 0, 1) as suppressed", t, t, t, t, t, t)).
		Joins("LEFT JOIN iac_policy_suppress AS e ON e.env_id = iac_policy_rel.env_id AND e.tpl_id = iac_policy_rel.tpl_id").
		Joins("LEFT JOIN iac_env ON iac_policy_rel.env_id = iac_env.id").
		Joins("LEFT JOIN iac_template ON iac_policy_rel.tpl_id = iac_template.id").
		Joins("LEFT JOIN iac_policy ON iac_policy_rel.group_id = iac_policy.group_id").
		Where("iac_policy.id = ?", id)
	return q
}

func SearchPolicyTpl(tx *db.Session, orgId models.Id, q string) *db.Session {
	query := tx.Table(models.Template{}.TableName())
	if orgId != "" {
		query = query.Where("org_id = ?", orgId)
	}

	if q != "" {
		query = query.WhereLike("name", q)
	}
	return query
}

func SearchPolicyEnv(dbSess *db.Session, orgId, projectId models.Id, q string) *db.Session {
	env := models.EnvDetail{}.TableName()
	query := dbSess.Table(env).
		Joins(fmt.Sprintf("left join %s as tpl on tpl.id = %s.tpl_id",
			models.Template{}.TableName(), env))
	if orgId != "" {
		query = query.Where(fmt.Sprintf("%s.org_id = ?", env), orgId)
	}

	if projectId != "" {
		query = query.Where(fmt.Sprintf("%s.project_id = ?", env), projectId)
	}

	if q != "" {
		query = query.WhereLike(fmt.Sprintf("%s.name", env), q)
	}
	return query.LazySelectAppend(fmt.Sprintf("%s.*", env)).
		LazySelectAppend("tpl.name as template_name,tpl.id as tpl_id,tpl.repo_addr")
}

func EnvOfPolicy(dbSess *db.Session, form *forms.EnvOfPolicyForm, orgId, projectId models.Id) *db.Session {
	pTable := models.Policy{}.TableName()
	query := dbSess.Table(pTable).Joins(fmt.Sprintf("left join %s as pg on pg.id = %s.group_id",
		models.PolicyGroup{}.TableName(), pTable)).LazySelectAppend("pg.name as group_name, pg.id as group_id")
	if form.GroupId != "" {
		query = query.Where(fmt.Sprintf("%s.group_id = ?", pTable), form.GroupId)
	}

	if form.Severity != "" {
		query = query.Where(fmt.Sprintf("%s.severity = ?", pTable), form.Severity)
	}

	if form.Q != "" {
		query = query.WhereLike(fmt.Sprintf("%s.name", pTable), form.Q)
	}

	query = query.
		Joins(fmt.Sprintf("left join %s as rel on rel.group_id = pg.id ", models.PolicyRel{}.TableName())).
		Joins(fmt.Sprintf("left join %s as env on env.id = rel.env_id", models.Env{}.TableName())).
		Where("rel.org_id = ? and rel.project_id = ? and rel.scope = ?", orgId, projectId, models.PolicyRelScopeEnv)

	return query.LazySelectAppend(fmt.Sprintf("env.name as env_name, %s.*", pTable))
}

func PolicyError(query *db.Session, policyId models.Id) *db.Session {
	return query.Model(models.PolicyResult{}).
		Select(fmt.Sprintf("if(%s.env_id='','template','env')as target_id,%s.*,%s.name as env_name,%s.name as template_name",
			models.PolicyResult{}.TableName(),
			models.PolicyResult{}.TableName(),
			models.Env{}.TableName(),
			models.Template{}.TableName(),
		)).
		Joins("LEFT JOIN iac_env ON iac_policy_result.env_id = iac_env.id").
		Joins("LEFT JOIN iac_template ON iac_policy_result.tpl_id = iac_template.id").
		Where("iac_policy_result.policy_id = ?", policyId).
		Where("iac_policy_result.status = ? OR iac_policy_result.status = ?",
			common.PolicyStatusSuppressed, common.PolicyStatusFailed, common.PolicyStatusViolated)

}

type PolicyScanSummary struct {
	Id     models.Id `json:"id"`
	Count  int       `json:"count"`
	Status string    `json:"status"`
}

// PolicySummary 获取策略/策略组/任务执行结果
func PolicySummary(query *db.Session, ids []models.Id, scope string) ([]*PolicyScanSummary, e.Error) {
	var key string
	switch scope {
	case consts.ScopePolicy:
		key = "policy_id"
	case consts.ScopePolicyGroup:
		key = "policy_group_id"
	case consts.ScopeTask:
		key = "task_id"
	}
	subQuery := query.Model(models.PolicyResult{}).Select("max(id)").
		Group(fmt.Sprintf("%s,env_id,tpl_id", key)).Where(fmt.Sprintf("%s in (?)", key), ids)
	q := query.Model(models.PolicyResult{}).Select(fmt.Sprintf("%s as id,count(*) as count,status", key)).
		Where("id in (?)", subQuery.Expr()).Group(fmt.Sprintf("%s,status", key))

	summary := make([]*PolicyScanSummary, 0)
	if err := q.Find(&summary); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, e.New(e.DBError, err)
	}

	return summary, nil
}

type ScanStatus struct {
	Date   string
	Count  int
	Status string
}

func GetPolicyScanStatus(query *db.Session, id models.Id, from time.Time, to time.Time, scope string) ([]*ScanStatus, e.Error) {
	key := ""
	switch scope {
	case consts.ScopePolicyGroup:
		key = "policy_group_id"
	case consts.ScopePolicy:
		key = "policy_id"
	}
	q := query.Model(models.PolicyResult{})
	q = q.Where(fmt.Sprintf("start_at >= ? and start_at < ? and %s = ?", key), from, to, id).
		Select("count(*) as count, date(start_at) as date, status").
		Group("date(start_at), status").
		Order("date(start_at)")

	scanStatus := make([]*ScanStatus, 0)
	if err := q.Find(&scanStatus); err != nil {
		if e.IsRecordNotFound(err) {
			return scanStatus, nil
		}
		return nil, e.New(e.DBError, err)
	}

	return scanStatus, nil
}

func GetPolicyScanByDate(query *db.Session, policyId models.Id, from time.Time, to time.Time) ([]*ScanStatus, e.Error) {
	q := query.Model(models.PolicyResult{})
	q = q.Where("start_at >= ? and start_at < ? and policy_id = ?", from, to, policyId).
		Where("status != 'pending'"). // 跳过pending状态
		Select("count(*) as count, date(start_at) as date").
		Group("date(start_at),tpl_id,env_id").
		Order("date(start_at)")

	scanStatus := make([]*ScanStatus, 0)
	if err := q.Find(&scanStatus); err != nil {
		if e.IsRecordNotFound(err) {
			return scanStatus, nil
		}
		return nil, e.New(e.DBError, err)
	}

	return scanStatus, nil
}

func SearchGroupOfPolicy(dbSess *db.Session, groupId models.Id, bind bool) *db.Session {
	query := dbSess.Table(models.Policy{}.TableName())
	if bind {
		query = query.Where("group_id = ? ", groupId)
	} else {
		query = query.Where("group_id = '' or group_id is null")
	}
	return query
}

func DeletePolicySuppress(tx *db.Session, id models.Id) e.Error {
	if _, err := tx.Where("policy_id = ?", id).Delete(models.PolicySuppress{}); err != nil {
		if e.IsRecordNotFound(err) {
			return nil
		}
		return e.New(e.DBError, err)
	}
	return nil
}

// PolicyTargetSummary 获取策略环境/云模板执行结果
func PolicyTargetSummary(query *db.Session, ids []models.Id, scope string) ([]*PolicyScanSummary, e.Error) {
	var key string
	switch scope {
	case consts.ScopeEnv:
		key = "env_id"
	case consts.ScopeTemplate:
		key = "tpl_id"
	}
	subQuery := query.Model(models.PolicyResult{}).Select("max(id)").
		Group(fmt.Sprintf("policy_id,%s", key)).Where(fmt.Sprintf("%s in (?)", key), ids)
	q := query.Model(models.PolicyResult{}).Select(fmt.Sprintf("%s as id,count(*) as count,status", key)).
		Where("id in (?)", subQuery.Expr()).Group(fmt.Sprintf("%s,status", key))

	summary := make([]*PolicyScanSummary, 0)
	if err := q.Find(&summary); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, e.New(e.DBError, err)
	}

	return summary, nil
}
