package apps

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"net/http"
	"strings"
)

type PolicySuppressResp struct {
	models.PolicySuppress
	TargetName string `json:"targetName"` // 检查目标
	Creator    string `json:"creator"`    // 操作人
}

func (PolicySuppressResp) TableName() string {
	return "s"
}

func SearchPolicySuppress(c *ctx.ServiceContext, form *forms.SearchPolicySuppressForm) (interface{}, e.Error) {
	query := services.QueryWithOrgId(c.DB(), c.OrgId)
	query = services.SearchPolicySuppress(query, form.Id)
	if form.SortField() == "" {
		query = query.Order(fmt.Sprintf("%s.created_at DESC", PolicySuppressResp{}.TableName()))
	}
	return getPage(query, form, PolicySuppressResp{})
}

func UpdatePolicySuppress(c *ctx.ServiceContext, form *forms.UpdatePolicySuppressForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("update policy suppress %s", form.Id))

	tx := services.QueryWithOrgId(c.Tx(), c.OrgId)
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 权限检查
	//ids := append(form.RmSourceIds, form.AddSourceIds...)
	for _, id := range form.AddSourceIds {
		if strings.HasPrefix(string(id), "env-") {
			env, err := services.GetEnvById(tx, id)
			if err != nil {
				_ = tx.Rollback()
				if err.Code() == e.EnvNotExists {
					return nil, e.New(err.Code(), err, http.StatusBadRequest)
				}
				return nil, e.New(e.DBError, err, http.StatusInternalServerError)
			}

			if !c.IsSuperAdmin && !services.UserHasOrgRole(c.UserId, env.OrgId, consts.OrgRoleAdmin) &&
				!services.UserHasProjectRole(c.UserId, env.OrgId, env.ProjectId, "") {
				_ = tx.Rollback()
				return nil, e.New(e.EnvNotExists, fmt.Errorf("cannot access env %s", id), http.StatusForbidden)
			}
		} else if strings.HasPrefix(string(id), "tpl-") {
			tpl, err := services.GetTemplateById(tx, id)
			if err != nil {
				_ = tx.Rollback()
				if err.Code() == e.TemplateNotExists {
					return nil, e.New(err.Code(), err, http.StatusBadRequest)
				}
				return nil, e.New(e.DBError, err, http.StatusInternalServerError)
			}
			if !c.IsSuperAdmin && !services.UserHasOrgRole(c.UserId, tpl.OrgId, "") {
				_ = tx.Rollback()
				return nil, e.New(e.TemplateNotExists, fmt.Errorf("cannot access tpl %s", id), http.StatusForbidden)
			}
		} else if strings.HasPrefix(string(id), "po-") {
			_, err := services.GetPolicyById(tx, id, c.OrgId)
			if err != nil {
				_ = tx.Rollback()
				if err.Code() == e.PolicyNotExist {
					return nil, e.New(err.Code(), err, http.StatusBadRequest)
				}
				return nil, e.New(e.DBError, err, http.StatusInternalServerError)
			}
		}
	}

	// 删除屏蔽记录
	//if err := services.DeletePolicySuppressIds(tx, form.Id, form.RmSourceIds); err != nil {
	//	_ = tx.Rollback()
	//	if err.Code() == e.PolicySuppressNotExist {
	//		return nil, e.New(e.DBError, err, http.StatusBadRequest)
	//	}
	//	return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	//}

	// 创新新的屏蔽记录
	var (
		sups []models.PolicySuppress
	)
	for _, id := range form.AddSourceIds {
		if strings.HasPrefix(string(id), "env-") {
			env, _ := services.GetEnvById(tx, id)
			sups = append(sups, models.PolicySuppress{
				CreatorId:  c.UserId,
				OrgId:      env.OrgId,
				ProjectId:  env.ProjectId,
				TargetId:   id,
				TargetType: consts.ScopeEnv,
				PolicyId:   form.Id,
				Type:       common.PolicySuppressTypeSource,
				Reason:     form.Reason,
			})
		} else if strings.HasPrefix(string(id), "tpl-") {
			tpl, _ := services.GetTemplateById(tx, id)
			sups = append(sups, models.PolicySuppress{
				CreatorId:  c.UserId,
				OrgId:      tpl.OrgId,
				TargetId:   id,
				TargetType: consts.ScopeTemplate,
				PolicyId:   form.Id,
				Type:       common.PolicySuppressTypeSource,
				Reason:     form.Reason,
			})
		} else if strings.HasPrefix(string(id), "po-") {
			// 一次只能提交一个策略禁用
			if len(form.AddSourceIds) > 1 {
				return nil, e.New(e.BadParam, fmt.Errorf("one policy id a time"), http.StatusBadRequest)
			}
			if form.Id != id {
				return nil, e.New(e.BadParam, fmt.Errorf("invalid policy id to disable"), http.StatusBadRequest)
			}
			po, _ := services.GetPolicyById(tx, id, c.OrgId)
			sups = append(sups, models.PolicySuppress{
				CreatorId:  c.UserId,
				TargetId:   id,
				TargetType: consts.ScopePolicy,
				PolicyId:   form.Id,
				Type:       common.PolicySuppressTypePolicy,
				Reason:     form.Reason,
			})
			// 禁用此策略在添加屏蔽的同时设置策略状态为禁用
			po.Enabled = false
			if _, err := tx.Save(po); err != nil {
				_ = tx.Rollback()
				return nil, e.New(e.DBError, err)
			}
		}
	}

	if er := models.CreateBatch(tx, sups); er != nil {
		_ = tx.Rollback()
		if e.IsDuplicate(er) {
			return nil, e.New(e.PolicySuppressAlreadyExist, er, http.StatusBadRequest)
		}
		return nil, e.New(e.DBError, er)
	}

	if err := tx.Commit(); err != nil {
		c.Logger().Errorf("error commit policy suppress, err %s", err)
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return sups, nil
}

func DeletePolicySuppress(c *ctx.ServiceContext, form *forms.DeletePolicySuppressForm) (interface{}, e.Error) {
	tx := services.QueryWithOrgId(c.Tx(), c.OrgId)
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	sup, err := services.GetPolicySuppressById(tx, form.SuppressId)
	if err != nil {
		c.Logger().Errorf("sup not exist, rollback, code %d", err.Code())
		_ = tx.Rollback()
		if err.Code() == e.PolicySuppressNotExist {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}
	if sup.TargetType == consts.ScopePolicy {
		_, err := services.PolicyEnable(tx, sup.TargetId, true, c.OrgId)
		if err != nil {
			_ = tx.Rollback()
			return nil, e.New(err.Code(), err, http.StatusInternalServerError)
		}
	}

	_, err = services.DeletePolicySuppress(tx, form.SuppressId)
	if err != nil {
		_ = tx.Rollback()
		if err.Code() == e.PolicySuppressNotExist {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	return nil, nil
}

type PolicySuppressSourceResp struct {
	TargetId   models.Id `json:"targetId" example:"env-c3lcrjxczjdywmk0go90"`   // 屏蔽源ID
	TargetType string    `json:"targetType" enums:"env,template" example:"env"` // 源类型：env环境, template云模板
	TargetName string    `json:"targetName" example:"测试环境"`                     // 名称
}

func (PolicySuppressSourceResp) TableName() string {
	return "iac_policy_suppress"
}

func SearchPolicySuppressSource(c *ctx.ServiceContext, form *forms.SearchPolicySuppressSourceForm) (interface{}, e.Error) {
	query := services.QueryWithOrgId(c.DB(), c.OrgId)
	policy, err := services.GetPolicyById(query, form.Id, c.OrgId)
	if err != nil {
		if err.Code() == e.PolicyNotExist {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		} else {
			return nil, e.New(err.Code(), err, http.StatusInternalServerError)
		}
	}
	query = services.SearchPolicySuppressSource(c.DB(), form, c.UserId, form.Id, policy.GroupId, c.OrgId)
	return getPage(query, form, PolicySuppressSourceResp{})
}
