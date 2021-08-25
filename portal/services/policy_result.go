package services

import (
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

// batchUpdatePolicyResult 批量更新结果
func batchUpdatePolicyResultStatus(tx *db.Session, policyResultIds []models.Id, status string) (int64, e.Error) {
	sql := fmt.Sprintf("UPDATE %s SET status = ? WHERE id IN (?)", models.PolicyResult{}.TableName())
	if affected, err := tx.Exec(sql, status, policyResultIds); err != nil {
		return affected, e.New(e.DBError, err)
	} else if int(affected) != len(policyResultIds) {
		return affected, e.New(e.DBError, err)
	} else {
		return affected, nil
	}
}

func InitScanResult(tx *db.Session, task models.Task) e.Error {
	// 1. 扫描类型：
	//      1. 环境
	//      2. 云模板
	var (
		policies      []models.Policy
		err           e.Error
		policyResults []*models.PolicyResult
	)

	// 根据扫描类型获取策略列表
	typ := "environment"
	switch typ {
	case "environment":
		policies, err = GetPoliciesByEnvId(tx, task.EnvId)
	case "template":
		if env, er := GetEnvById(tx, task.EnvId); er != nil {
			return er
		} else {
			policies, err = GetPoliciesByTemplateId(tx, env.TplId)
		}
	default:
		return e.New(e.InternalError, fmt.Errorf("not support scan type"))
	}
	if err != nil {
		return err
	}

	// 批量创建
	for _, policy := range policies {
		policyResults = append(policyResults, &models.PolicyResult{
			OrgId:     task.OrgId,
			ProjectId: task.ProjectId,
			TplId:     task.TplId,
			EnvId:     task.EnvId,
			TaskId:    task.Id,

			PolicyId:      policy.Id,
			PolicyGroupId: policy.GroupId,

			StartAt: models.Time(time.Now()),
			Status:  "pending", // 设置结果为等待扫描状态
		})
	}

	if er := models.CreateBatch(tx, policyResults); er != nil {
		return e.New(e.DBError, er)
	}

	return nil
}

// UpdateScanResult 根据 terrascan 扫描结果批量更新
func UpdateScanResult(tx *db.Session, task models.Task, result TsResult) e.Error {
	var (
		policyResults []*models.PolicyResult
	)
	for _, r := range result.Violations {
		if policyResult, err := GetPolicyResultById(tx, task.Id, models.Id(r.RuleId)); err != nil {
			return err
		} else {
			policyResult.Status = "violated"
			policyResults = append(policyResults, policyResult)
		}
	}
	for _, r := range result.PassedRules {
		if policyResult, err := GetPolicyResultById(tx, task.Id, models.Id(r.RuleId)); err != nil {
			return err
		} else {
			policyResult.Status = "passed"
			policyResults = append(policyResults, policyResult)
		}
	}
	for _, r := range policyResults {
		if err := models.Save(tx, r); err != nil {
			return e.New(e.DBError, fmt.Errorf("save scan result"))
		}
	}

	if err := finishScanResult(tx, &task); err != nil {
		return err
	}
	return nil
}

// finishScanResult 更新状态未知的策略扫描结果
// 当存在相同名称当策略时，扫描结果可能不包含在结果集里面，这些策略扫描结果一律标记为 failed
func finishScanResult(tx *db.Session, task *models.Task) e.Error {
	sql := fmt.Sprintf("UPDATE %s SET status = 'failed' WHERE task_id = ? AND status = 'pending'",
		models.PolicyResult{}.TableName())
	if _, err := tx.Exec(sql, task.Id); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

/*
获取扫描结果

查询最新扫描结果：
按 org, policy id, policy group id，按 start at 取最新一条记录
根据范围做 distinct:
1. 查看策略
2. 查看策略组
3. 查看环境
4. 查看云模板

1. 策略页面获取当前策略扫描状态
2. 策略组页面获取当前策略组扫描状态
	是否扫描中：
		检查是否存在 result.status = pending
	是否违反：
		检查是否存在 result.status = violated
	是否存在错误：

'passed','violated','suppressed','pending','failed'
*/

func GetPolicyResultByPolicyId(query *db.Session, policyId models.Id) error {
	return nil
}
