package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/runner"
	"fmt"
	"strconv"
	"strings"
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
		meta := map[string]interface{}{
			"name":          policy.Name,
			"file":          "policy.rego",
			"policy_type":   policy.PolicyType,
			"resource_type": policy.ResourceType,
			"severity":      strings.ToUpper(policy.Severity),
			"reference_id":  policy.ReferenceId,
			"category":      policy.Category,
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
	// TODO: 处理策略关联
	var policyIds []models.Id
	if err := query.Model(models.Policy{}).Pluck("id", &policyIds); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return policyIds, nil
}

func GetPolicyGroupsByEnvId(query *db.Session, envId models.Id) ([]models.PolicyGroup, error) {
	return nil, nil
}

func GetPoliciesByGroupIds(query *db.Session, groupId ...models.Id) ([]models.Policy, error) {
	return nil, nil
}

func GetPoliciesByEnvId(query *db.Session, envId models.Id) ([]models.Policy, e.Error) {
	return nil, nil
}

func GetPoliciesByTemplateId(query *db.Session, envId models.Id) ([]models.Policy, e.Error) {
	return nil, nil
}

func BindPolicyGroupWithEnv(tx *db.Session, groupIds []models.Id, envId models.Id) e.Error {
	return nil
}

func BindPolicyGroupWithTemplate(tx *db.Session, groupIds []models.Id, tplId models.Id) e.Error {
	return nil
}
