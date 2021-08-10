package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

func CreatePolicy(tx *db.Session, policy *models.Policy) (*models.Policy, e.Error) {
	policy.Id = policy.GetId()
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
	if err := query.Model(models.Policy{}).Where("reference_id LIKE ?%", "iac_"+typ).Order("reference_id ASC").Last(&po); err != nil && !e.IsRecordNotFound(err) {
		return "", nil
	}

	// internal or public
	scope := "public"

	return fmt.Sprintf("%s_%s_%s_%d", "iac_", typ, scope, lastId+1), nil
}
