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

	query = query.Debug()
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

func suppressedQuery(query *db.Session, envId models.Id, tplId models.Id) *db.Session {
	q := query.Select("iac_policy.id").Table(models.PolicyRel{}.TableName()).
		Joins("join iac_policy on iac_policy.group_id = iac_policy_rel.group_id").
		Where("iac_policy.enabled = 1").
		Joins("join iac_policy_group on iac_policy_group.id = iac_policy_rel.group_id").
		Where("iac_policy_group.enabled = 1")

	if envId != "" {
		q = q.Where("iac_policy_rel.env_id = ?", envId)
	} else if tplId != "" {
		q = q.Where("iac_policy_rel.tpl_id = ?", tplId)
	}

	suppressQuery := query.Model(models.PolicySuppress{}).Select("policy_id")
	if envId != "" {
		suppressQuery = suppressQuery.Where("target_type = 'env' AND env_id = ?", envId)
		q = q.Where("iac_policy.id not in (?)", suppressQuery.Expr())
	} else if tplId != "" {
		suppressQuery = suppressQuery.Where("target_type = 'template' AND tpl_id = ?", tplId)
		q = q.Where("iac_policy.id not in (?)", suppressQuery.Expr())
	}

	enableQuery := query.Model(models.PolicyRel{}).Where("iac_policy_rel.group_id = '' and iac_policy_rel.enabled = 1")
	if envId != "" {
		enableQuery = enableQuery.Select("env_id").Where("env_id = ?", envId)
		q = q.Where("iac_policy_rel.env_id in (?)", enableQuery.Expr())
	} else if tplId != "" {
		enableQuery = enableQuery.Select("tpl_id").Where("tpl_id = ?", tplId)
		q = q.Where("iac_policy_rel.tpl_id in (?)", enableQuery.Expr())
	}

	return q
}

// GetPoliciesByEnvId 查询环境关联的策略，排除已经禁用的策略/策略组/策略屏蔽
func GetPoliciesByEnvId(query *db.Session, envId models.Id) ([]models.Policy, e.Error) {
	subQuery := suppressedQuery(query, envId, "")
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
	subQuery := suppressedQuery(query, "", tplId)
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

func SearchPolicyTpl(tx *db.Session, orgId, tplId models.Id, q string) *db.Session {
	query := tx.Table("iac_template AS tpl")
	if orgId != "" {
		query = query.Where("tpl.org_id = ?", orgId)
	}
	if tplId != "" {
		query = query.Where("tpl.id = ?", tplId)
	}
	if q != "" {
		query = query.WhereLike("tpl.name", q)
	}
	query = query.Joins("LEFT JOIN iac_scan_task AS task ON task.id = tpl.last_scan_task_id")
	return query.LazySelect("tpl.*, task.status AS scan_task_status").
		Joins("LEFT JOIN iac_policy_rel on iac_policy_rel.tpl_id = tpl.id and iac_policy_rel.group_id = ''").
		LazySelectAppend("iac_policy_rel.enabled")
}

func SearchPolicyEnv(dbSess *db.Session, orgId, projectId, envId models.Id, q string) *db.Session {
	envTable := models.Env{}.TableName()
	query := dbSess.Table(envTable)
	if orgId != "" {
		query = query.Where(fmt.Sprintf("%s.org_id = ?", envTable), orgId)
	}
	if projectId != "" {
		query = query.Where(fmt.Sprintf("%s.project_id = ?", envTable), projectId)
	}
	if envId != "" {
		query = query.Where(fmt.Sprintf("%s.id = ?", envTable), envId)
	}

	if q != "" {
		query = query.WhereLike(fmt.Sprintf("%s.name", envTable), q)
	}

	query = query.Joins(fmt.Sprintf("LEFT JOIN %s AS tpl ON tpl.id = %s.tpl_id",
		models.Template{}.TableName(), envTable))
	query = query.Joins(fmt.Sprintf("LEFT JOIN %s AS task ON task.id = %s.last_scan_task_id",
		models.ScanTask{}.TableName(), envTable))

	return query.
		LazySelectAppend(fmt.Sprintf("%s.*", envTable)).
		LazySelectAppend("tpl.name AS template_name, tpl.id AS tpl_id, tpl.repo_addr AS repo_addr").
		LazySelectAppend("task.status AS scan_task_status").
		Joins("LEFT JOIN iac_policy_rel on iac_policy_rel.env_id = iac_env.id and iac_policy_rel.group_id = ''").
		LazySelectAppend("iac_policy_rel.enabled")
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

func TplOfPolicy(dbSess *db.Session, form *forms.TplOfPolicyForm, orgId, projectId models.Id) *db.Session {
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
		Joins(fmt.Sprintf("left join %s as tpl on tpl.id = rel.tpl_id", models.Template{}.TableName())).
		Where("rel.org_id = ? and rel.project_id = ? and rel.scope = ?", orgId, projectId, models.PolicyRelScopeTpl)

	return query.LazySelectAppend(fmt.Sprintf("tpl.name as tpl_name, %s.*", pTable))
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

type ScanStatusByTarget struct {
	ID     string
	Count  int
	Status string
	Name   string
}

func GetPolicyScanByTarget(query *db.Session, policyId models.Id, from time.Time, to time.Time) ([]*ScanStatusByTarget, e.Error) {
	groupQuery := query.Model(models.PolicyResult{})
	groupQuery = groupQuery.Where("start_at >= ? and start_at < ? and policy_id = ?", from, to, policyId).
		Where("status != 'pending'"). // 跳过pending状态
		Select("count(*) as count, tpl_id, env_id").
		Group("tpl_id,env_id")

	q := query.Table("(?) as r", groupQuery.Expr()).
		Select("r.*,if(r.env_id = '', iac_template.name, iac_env.name) as name").
		Joins("left join iac_env on iac_env.id = r.env_id").
		Joins("left join iac_template on iac_template.id = r.tpl_id")

	scanStatus := make([]*ScanStatusByTarget, 0)
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

func SubQueryUserEnvIds(query *db.Session, userId models.Id) *db.Session {
	q := query.Model(models.Env{}).Select("id")

	// 系统管理员
	if UserIsSuperAdmin(query, userId) {
		return q
	}

	// 组织管理员相关项目
	var orgAdminIds []models.Id
	userOrgs := getUserOrgs(userId)
	orgAdminProjectQuery := query.Model(models.Project{}).Select("id")
	for _, userOrg := range userOrgs {
		if UserHasOrgRole(userId, userOrg.OrgId, consts.OrgRoleAdmin) {
			orgAdminIds = append(orgAdminIds, userOrg.OrgId)
		}
	}

	// 普通成员相关项目
	var projectIds []models.Id
	userProjects := getUserProjects(userId)
	for _, userProject := range userProjects {
		projectIds = append(projectIds, userProject.ProjectId)
	}

	if len(orgAdminIds) > 0 && len(projectIds) > 0 {
		return q.Where("org_id in (?) OR project_id in (?)", orgAdminProjectQuery, projectIds)
	} else if len(orgAdminIds) > 0 {
		return q.Where("org_id in (?)", orgAdminProjectQuery)
	} else if len(projectIds) > 0 {
		return q.Where("project_id in (?)", projectIds)
	} else {
		return q.Where("1 = 0")
	}
}

func SubQueryUserTemplateIds(query *db.Session, userId models.Id) *db.Session {
	q := query.Model(models.Template{}).Select("id")

	// 平台管理员
	if UserIsSuperAdmin(query, userId) {
		return q
	}

	// 组织内用户
	userOrgs := getUserOrgs(userId)
	var orgIds []models.Id
	for _, userOrg := range userOrgs {
		orgIds = append(orgIds, userOrg.OrgId)
	}
	if len(orgIds) == 0 {
		return q.Where("1 = 0")
	}

	return q.Where("org_id in (?)", orgIds)
}

func PolicyEnable(tx *db.Session, policyId models.Id, enabled bool) (*models.Policy, e.Error) {
	policy, err := GetPolicyById(tx, policyId)
	if err != nil {
		return nil, err
	}
	policy.Enabled = enabled
	_, er := tx.Save(policy)
	if er != nil {
		return nil, e.New(e.DBError, fmt.Errorf("save policy enable error, id %s", policyId))
	}

	return policy, nil
}

type ScanStatusGroupBy struct {
	Id       models.Id
	Name     string
	Count    int
	Status   string
	Severity string
}

func GetPolicyStatusByPolicy(query *db.Session, from time.Time, to time.Time, status string) ([]*ScanStatusGroupBy, e.Error) {
	groupQuery := query.Model(models.PolicyResult{})
	groupQuery = groupQuery.Where("start_at >= ? and start_at < ?", from, to).
		Select("count(*) as count, policy_id as id, status").
		Group("policy_id,status").
		Order("count desc")

	q := query.Select("r.*,iac_policy.name,iac_policy.severity").Table("(?) as r", groupQuery.Expr()).
		Joins("left join iac_policy on iac_policy.id = r.id")

	if status != "" {
		q = q.Where("r.status = ?", status)
	}

	scanStatus := make([]*ScanStatusGroupBy, 0)
	if err := q.Find(&scanStatus); err != nil {
		if e.IsRecordNotFound(err) {
			return scanStatus, nil
		}
		return nil, e.New(e.DBError, err)
	}

	return scanStatus, nil
}

func GetPolicyStatusByPolicyGroup(query *db.Session, from time.Time, to time.Time, status string) ([]*ScanStatusGroupBy, e.Error) {
	groupQuery := query.Model(models.PolicyResult{})
	groupQuery = groupQuery.Where("start_at >= ? and start_at < ?", from, to).
		Select("count(*) as count, policy_group_id as id, status").
		Group("policy_group_id,status").
		Order("count desc")

	q := query.Select("r.*,iac_policy_group.name").Table("(?) as r", groupQuery.Expr()).
		Joins("left join iac_policy_group on iac_policy_group.id = r.id")

	if status != "" {
		q = q.Where("r.status = ?", status)
	}

	scanStatus := make([]*ScanStatusGroupBy, 0)
	if err := q.Find(&scanStatus); err != nil {
		if e.IsRecordNotFound(err) {
			return scanStatus, nil
		}
		return nil, e.New(e.DBError, err)
	}

	return scanStatus, nil
}
