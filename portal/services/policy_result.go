// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package services

import (
	"cloudiac/common"
	"cloudiac/policy"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
	"time"
)

func GetPolicyResultById(query *db.Session, taskId models.Id, policyId models.Id) (*models.PolicyResult, e.Error) {
	result := models.PolicyResult{}
	if err := query.Model(models.PolicyResult{}).Where("task_id = ? AND policy_id = ?", taskId, policyId).First(&result); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.PolicyResultNotExist, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &result, nil
}

func GetPoliciesByTaskId(query *db.Session, taskId models.Id) ([]*models.Policy, e.Error) {
	var policies []*models.Policy
	resultQuery := query.Model(models.PolicyResult{}).Where("task_id = ?", taskId).Select("policy_id")
	if err := query.Model(models.Policy{}).
		Where("id in (?)", resultQuery.Expr()).
		Find(&policies); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return policies, nil
}

// InitScanResult 初始化扫描结果
func InitScanResult(tx *db.Session, task *models.ScanTask) e.Error {
	var (
		validPolicies, suppressedPolicies []models.Policy
		policyResults                     []*models.PolicyResult
		err                               e.Error
	)

	if validPolicies, suppressedPolicies, err = GetValidPolicies(tx, task.TplId, task.EnvId); err != nil {
		return err
	}

	if len(validPolicies) == 0 && len(suppressedPolicies) == 0 {
		return nil
	}

	// 批量创建
	for _, policy := range validPolicies {
		policyResults = append(policyResults, &models.PolicyResult{
			OrgId:     task.OrgId,
			ProjectId: task.ProjectId,
			TplId:     task.TplId,
			EnvId:     task.EnvId,
			TaskId:    task.Id,

			PolicyId:      policy.Id,
			PolicyGroupId: policy.GroupId,

			StartAt: models.Time(time.Now()),
			Status:  common.TaskStepPending,
			Violation: models.Violation{
				Severity: policy.Severity,
			},
		})
	}
	for _, policy := range suppressedPolicies {
		policyResults = append(policyResults, &models.PolicyResult{
			OrgId:     task.OrgId,
			ProjectId: task.ProjectId,
			TplId:     task.TplId,
			EnvId:     task.EnvId,
			TaskId:    task.Id,

			PolicyId:      policy.Id,
			PolicyGroupId: policy.GroupId,

			StartAt: models.Time(time.Now()),
			Status:  common.PolicyStatusSuppressed,
			Violation: models.Violation{
				Severity: policy.Severity,
			},
		})
	}

	if er := models.CreateBatch(tx, policyResults); er != nil {
		return e.New(e.DBError, er)
	}

	return nil
}

// UpdateScanResult 根据 terrascan 扫描结果批量更新
func UpdateScanResult(tx *db.Session, task models.Tasker, result policy.TsResult, policyStatus string) e.Error {

	var (
		policyResults []*models.PolicyResult
	)
	for _, r := range result.Violations {
		if policyResult, err := GetPolicyResultById(tx, task.GetId(), models.Id(r.RuleId)); err != nil {
			return err
		} else {
			policyResult.Status = "violated"
			policyResult.Line = r.Line
			policyResult.Source = r.Source
			policyResult.PlanRoot = r.PlanRoot
			policyResult.ModuleName = r.ModuleName
			policyResult.File = r.File
			policyResult.Violation = models.Violation{
				RuleName:     r.RuleName,
				Description:  r.Description,
				RuleId:       r.RuleId,
				Severity:     r.Severity,
				Category:     r.Category,
				Comment:      r.Comment,
				ResourceName: r.ResourceName,
				ResourceType: r.ResourceType,
				ModuleName:   r.ModuleName,
				File:         r.File,
				PlanRoot:     r.PlanRoot,
				Line:         r.Line,
				Source:       r.Source,
			}
			policyResults = append(policyResults, policyResult)
		}
	}
	for _, r := range result.PassedRules {
		if policyResult, err := GetPolicyResultById(tx, task.GetId(), models.Id(r.RuleId)); err != nil {
			return err
		} else {
			policyResult.Status = common.PolicyStatusPassed
			policyResults = append(policyResults, policyResult)
		}
	}
	for _, r := range policyResults {
		if err := models.Save(tx, r); err != nil {
			return e.New(e.DBError, fmt.Errorf("save scan result"))
		}
	}

	message := "policy skipped"
	status := common.PolicyStatusPassed
	if err := finishPendingScanResult(tx, task, message, status); err != nil {
		return err
	}
	return nil
}

// finishScanResult 更新状态未知的策略扫描结果
func finishPendingScanResult(tx *db.Session, task models.Tasker, message string, status string) e.Error {
	table := models.PolicyResult{}.TableName()
	sql := fmt.Sprintf("UPDATE %s SET status = ?, message = ? WHERE task_id = ? AND status = ?", table)
	args := []interface{}{status, message, task.GetId(), common.PolicyStatusPending}
	if _, err := tx.Exec(sql, args...); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

// CleanScanResult 任务失败的时候清除扫描结果
func CleanScanResult(tx *db.Session, task models.Tasker) e.Error {
	if _, err := tx.Where("task_id = ?", task.GetId()).
		Delete(models.PolicyResult{}); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

func GetPolicyGroupScanTasks(query *db.Session, policyGroupId, orgId models.Id) *db.Session {
	t := models.PolicyResult{}.TableName()
	subQuery := query.Model(models.PolicyResult{}).
		Select(fmt.Sprintf("%s.task_id,%s.policy_group_id,%s.env_id,%s.tpl_id", t, t, t, t)).
		Where("iac_policy_result.policy_group_id = ?", policyGroupId).
		Where("iac_policy_result.org_id = ?", orgId).
		Group("iac_policy_result.task_id,iac_policy_result.env_id,iac_policy_result.tpl_id,iac_policy_result.policy_group_id")

	q := query.Model(models.ScanTask{}).
		Joins("LEFT JOIN (?) AS r ON r.task_id = iac_scan_task.id", subQuery.Expr()).
		Where("r.policy_group_id = ?", policyGroupId).
		LazySelectAppend("iac_scan_task.*,r.*")

	q = q.Joins("LEFT JOIN iac_env ON r.env_id = iac_env.id").
		LazySelectAppend("if(r.env_id='','template','env')as target_type").
		Joins("LEFT JOIN iac_template ON r.tpl_id = iac_template.id").
		LazySelectAppend("if(r.env_id = '',iac_template.name,iac_env.name) as target_name")

	// 创建者
	q = q.Joins("left join iac_user as u on u.id = iac_scan_task.creator_id").
		LazySelectAppend("u.name as creator")
	// 组织
	q = q.Joins("left join iac_org as o on o.id = iac_scan_task.org_id").
		LazySelectAppend("o.name as org_name")
	// 项目
	q = q.Joins("left join iac_project as p on p.id = iac_scan_task.project_id").
		LazySelectAppend("p.name as project_name")
	return q
}

func QueryPolicyResult(query *db.Session, taskId models.Id) *db.Session {
	q := query.Model(models.PolicyResult{}).Where("task_id = ?", taskId)

	// 策略信息
	q = q.Joins("left join iac_policy as p on p.id = iac_policy_result.policy_id").
		LazySelectAppend("p.name as policy_name, p.fix_suggestion,iac_policy_result.*,p.rego")
	// 策略组信息
	q = q.Joins("left join iac_policy_group as g on g.id = iac_policy_result.policy_group_id").
		LazySelectAppend("g.name as policy_group_name,iac_policy_result.*")

	return q
}

//GetMirrorScanTask 查找部署任务对应的扫描任务
func GetMirrorScanTask(query *db.Session, taskId models.Id) (*models.ScanTask, e.Error) {
	t := models.ScanTask{}
	if err := query.Where("mirror = 1 AND mirror_task_id = ?", taskId).First(&t); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.TaskNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &t, nil
}

func FilterSuppressPolicies(query *db.Session, policies []models.Policy, targetId models.Id, scope string) (
	validPolicies []models.Policy, suppressedPolicies []models.Policy, err e.Error) {
	var (
		policyIds []models.Id
	)
	for _, policy := range policies {
		policyIds = append(policyIds, policy.Id)
	}

	suppressQuery := query.Table(fmt.Sprintf("%s as s", models.PolicySuppress{}.TableName())).
		Select("policy_id").
		Where("s.policy_id in (?)", policyIds).
		Where("(s.target_type = 'policy') OR (s.target_id = ? AND s.target_type = ?)", targetId, scope).
		Group("policy_id")

	// 搜索策略屏蔽 或者 来源屏蔽
	q := query.Model(models.Policy{}).Where("id in (?)", suppressQuery.Expr())
	if er := q.Find(&suppressedPolicies); er != nil {
		if e.IsRecordNotFound(er) {
			return policies, nil, nil
		}
		return policies, nil, e.New(e.DBError, er)
	}

	if len(suppressedPolicies) == 0 {
		return policies, nil, nil
	}

	// 区分有效策略和屏蔽策略
	suppressPolicyMap := make(map[models.Id]models.Policy)
	for idx, policy := range suppressedPolicies {
		suppressPolicyMap[policy.Id] = suppressedPolicies[idx]
	}
	for idx, policy := range policies {
		if _, ok := suppressPolicyMap[policy.Id]; !ok {
			validPolicies = append(validPolicies, policies[idx])
		}
	}

	return validPolicies, suppressedPolicies, nil
}

func MergePolicies(policies1, policies2 []models.Policy) (mergedPolicies []models.Policy) {
	policiesMap := make(map[models.Id]models.Policy)
	for idx, policy := range policies1 {
		policiesMap[policy.Id] = policies1[idx]
	}
	for idx, policy := range policies2 {
		policiesMap[policy.Id] = policies2[idx]
	}
	for id, _ := range policiesMap {
		mergedPolicies = append(mergedPolicies, policiesMap[id])
	}

	return
}
