package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
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

	if err := tx.Commit(); err != nil {
		c.Logger().Errorf("error commit policy group, err %s", err)
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return group, nil
}

type PolicyGroupResp struct {
	models.PolicyGroup
	PolicyCount uint `json:"policyCount" form:"policyCount" `
}

// SearchPolicyGroup 查询策略组列表
func SearchPolicyGroup(c *ctx.ServiceContext, form *forms.SearchPolicyGroupForm) (interface{}, e.Error) {
	query := services.SearchPolicyGroup(c.DB().Debug(), c.OrgId, form.Q)
	pg := make([]PolicyGroupResp, 0)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	if err := p.Scan(&pg); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return pg, nil
}

// UpdatePolicyGroup 修改策略组
func UpdatePolicyGroup(c *ctx.ServiceContext, form *forms.UpdatePolicyGroupForm) (interface{}, e.Error) {
	attr := models.Attrs{}
	if form.HasKey("name") {
		attr["name"] = form.Name
	}

	if form.HasKey("description") {
		attr["description"] = form.Description
	}

	if form.HasKey("status") {
		attr["status"] = form.Status
	}

	pg := models.PolicyGroup{}
	pg.Id = form.Id
	if err := services.UpdatePolicyGroup(c.DB(), &pg, attr); err != nil {
		return nil, err
	}
	return nil, nil
}

// DeletePolicyGroup 删除策略组
func DeletePolicyGroup(c *ctx.ServiceContext, form *forms.DeletePolicyGroupForm) (interface{}, e.Error) {
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 解除策略与策略组的关系
	// todo 手动解除关联的策略才可以删除 会好一些？
	if err := services.RemovePoliciesGroupRelation(tx, form.Id); err != nil {
		return nil, err
	}

	// 删除策略组
	if err := services.DeletePolicyGroup(tx, form.Id); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return nil, nil
}

// DetailPolicyGroup 查询策略组详情
func DetailPolicyGroup(c *ctx.ServiceContext, form *forms.DetailPolicyGroupForm) (interface{}, e.Error) {
	return services.DetailPolicyGroup(c.DB(), form.Id)
}

// OpPolicyAndPolicyGroupRel 创建和修改策略和策略组的关系
func OpPolicyAndPolicyGroupRel(c *ctx.ServiceContext, form *forms.OpnPolicyAndPolicyGroupRelForm) (interface{}, e.Error) {
	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if form.HasKey("addPolicyIds") && len(form.AddPolicyIds) > 0 {
		for _, policyId := range form.AddPolicyIds {
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
		if affected, err := services.UpdatePolicy(tx.Where("id in (?)", form.AddPolicyIds),
			&models.Policy{}, models.Attrs{"group_id": form.PolicyGroupId}); err != nil {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		} else if int(affected) != len(form.AddPolicyIds) {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		}
	}

	if form.HasKey("rmPolicyIds") && len(form.RmPolicyIds) > 0 {
		for _, policyId := range form.RmPolicyIds {
			policy, err := services.GetPolicyById(tx, models.Id(policyId))
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
			// 一个策略只能关联到一个策略组
			if policy.GroupId != form.PolicyGroupId {
				_ = tx.Rollback()
				return nil, e.New(e.PolicyBelongedToAnotherGroup, fmt.Errorf("policy belonged to another group"), http.StatusBadRequest)
			}
		}
		// 批量更新
		if affected, err := services.UpdatePolicy(tx.Where("id in (?)", form.RmPolicyIds),
			&models.Policy{}, models.Attrs{"group_id": ""}); err != nil {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		} else if int(affected) != len(form.RmPolicyIds) {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, err, http.StatusInternalServerError)
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}
	return nil, nil
}
