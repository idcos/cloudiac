package services

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"fmt"
)

func SearchPolicySuppress(query *db.Session, id, orgId models.Id) *db.Session {
	q := query.Table(fmt.Sprintf("%s as s", models.PolicySuppress{}.TableName())).
		LazySelect("s.*")

	q = q.Joins("LEFT JOIN iac_env AS e ON s.target_id = e.id AND s.target_type = 'env'").
		Joins("LEFT JOIN iac_template AS t ON s.target_id = t.id AND s.target_type = 'template'").
		Joins("LEFT JOIN iac_policy AS p ON s.policy_id = p.id AND s.target_type = 'policy'").
		LazySelectAppend(`case 
when s.target_type = 'env' then e.name 
when s.target_type = 'template' then t.name 
when s.target_type = 'policy' then p.name
end as target_name`).
		Where("s.policy_id = ?", id).
		Where("s.org_id = ?", orgId).
		Joins("LEFT JOIN iac_user AS u ON s.creator_id = u.id").
		LazySelectAppend("u.name as creator")

	return q
}

func DeletePolicySuppress(dbSess *db.Session, id models.Id) (interface{}, e.Error) {
	if cnt, err := dbSess.Where("id = ?", id).Delete(&models.PolicySuppress{}); err != nil {
		if cnt != 1 {
			return nil, e.New(e.PolicySuppressNotExist, fmt.Errorf("policy suppress not exist, id: %s", id))
		}
		return nil, e.New(e.DBError, err)
	}

	return nil, nil
}

func GetPolicySuppressById(query *db.Session, id models.Id) (*models.PolicySuppress, e.Error) {
	sup := models.PolicySuppress{}
	if err := query.Model(models.PolicySuppress{}).Where("id = ?", id).First(&sup); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.PolicySuppressNotExist, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &sup, nil
}

func SearchPolicySuppressSource(query *db.Session, form *forms.SearchPolicySuppressSourceForm, userId, policyId, policyGroupId, orgId models.Id) *db.Session {
	subQueryPolicyGroupRel := query.Model(models.PolicyRel{}).Where("group_id = ?", policyGroupId)

	envQuery := query.Model(models.Env{}).
		Select("iac_env.id as target_id, iac_env.name as target_name, 'env' as target_type").
		Where("id in (?)", SubQueryUserEnvIds(query, userId).Expr()).
		Where("id in (?)", subQueryPolicyGroupRel.Select("env_id").Expr()).
		Where("org_id = ?", orgId)

	templateQuery := query.Model(models.Template{}).
		Select("iac_template.id as target_id, iac_template.name as target_name, 'template' as target_type").
		Where("id in (?)", SubQueryUserTemplateIds(query, userId).Expr()).
		Where("id in (?)", subQueryPolicyGroupRel.Select("tpl_id").Expr()).
		Where("org_id = ?", orgId)

	suppressQuery := query.Model(models.PolicySuppress{}).Where("policy_id = ?", policyId).
		Select("target_id")

	q := query.Raw(fmt.Sprintf("select r.* from ((?) union (?)) as r where r.target_id not in (?) %s", form.OrderBy()),
		envQuery.Expr(), templateQuery.Expr(), suppressQuery.Expr())

	return q
}

// GetPolicySuppressByPolicyId 获取策略禁用/启用屏蔽记录
func GetPolicySuppressByPolicyId(query *db.Session, id models.Id) (*models.PolicySuppress, e.Error) {
	sup := models.PolicySuppress{}
	if err := query.Model(models.PolicySuppress{}).Where("target_id = ? and type = 'policy'", id).First(&sup); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.PolicySuppressNotExist, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &sup, nil
}
