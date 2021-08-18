package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"net/http"
)

// CreatePolicyGroup 创建策略组
func CreatePolicyGroup(c *ctx.ServiceContext, form *forms.CreatePolicyGroupForm) (*models.PolicyGroup, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create policy group %s", form.Name))

	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	g := models.PolicyGroup{
		OrgId:       c.OrgId,
		Name:        form.Name,
		Description: form.Description,
	}

	group, err := services.CreatePolicyGroup(tx, &g)
	if err != nil && err.Code() == e.PolicyGroupAlreadyExist {
		_ = tx.Rollback()
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error creating policy group, err %s", err)
		_ = tx.Rollback()
		return nil, e.AutoNew(err, e.DBError)
	}

	if form.HasKey("policyIds") {
		for _, policyId := range form.PolicyIds {
			policy, err := services.GetPolicyById(tx, models.Id(policyId))
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
			// 一个策略只能关联到一个策略组
			if policy.GroupId != "" {
				_ = tx.Rollback()
				return nil, e.New(e.PolicyBelongedToAnotherGroup, fmt.Errorf("policy belonged to another group"), http.StatusBadRequest)
			}
		}
		// 批量更新
		sql := fmt.Sprintf("UPDATE %s SET group_id = ? WHERE id IN (?)", models.Policy{}.TableName())
		if affected, err := tx.Exec(sql, group.Id, form.PolicyIds); err != nil {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		} else if int(affected) != len(form.PolicyIds) {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		}
	}

	if err := tx.Commit(); err != nil {
		c.Logger().Errorf("error commit policy group, err %s", err)
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return group, nil
}
