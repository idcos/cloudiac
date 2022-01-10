// Copyright 2021 CloudJ Company Limited. All rights reserved.

package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils/logs"
	"fmt"
	"net/http"
)

// UpdatePolicyRel 创建/更新策略关系
func UpdatePolicyRel(tx *db.Session, form *forms.UpdatePolicyRelForm) ([]models.PolicyRel, e.Error) {
	logs.Get().Info("action", fmt.Sprintf("create policy relation %s %s", form.Scope, form.Id))

	var (
		env  *models.Env
		tpl  *models.Template
		rels []models.PolicyRel
		err  e.Error
	)
	if form.Scope == consts.ScopeEnv {
		env, err = services.GetEnvById(tx, form.Id)
		if err != nil {
			_ = tx.Rollback()
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
	} else {
		tpl, err = services.GetTemplateById(tx, form.Id)
		if err != nil {
			_ = tx.Rollback()
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
	}

	// 删除原有关联关系
	if err := services.DeletePolicyGroupRel(tx, form.Id, form.Scope); err != nil {
		return nil, e.New(e.DBError, err, http.StatusInternalServerError)
	}

	// 创新新的关联关系
	for _, groupId := range form.PolicyGroupIds {
		group, err := services.GetPolicyGroupById(tx, groupId)
		if err != nil {
			_ = tx.Rollback()
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
		if env != nil {
			rels = append(rels, models.PolicyRel{
				OrgId:     env.OrgId,
				ProjectId: env.ProjectId,
				GroupId:   group.Id,
				EnvId:     env.Id,
				Scope:     consts.ScopeEnv,
			})
		} else {
			rels = append(rels, models.PolicyRel{
				OrgId:   tpl.OrgId,
				GroupId: group.Id,
				TplId:   tpl.Id,
				Scope:   models.PolicyRelScopeTpl,
			})
		}
	}

	if len(rels) > 0 {
		if er := models.CreateBatch(tx, rels); er != nil {
			_ = tx.Rollback()
			return nil, e.New(e.DBError, er)
		}
	}

	if err := tx.Commit(); err != nil {
		logs.Get().Errorf("error commit policy relations, err %s", err)
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return rels, nil
}

// EnablePolicyScanRel 启用环境/云模板扫描
func EnablePolicyScanRel(c *ctx.ServiceContext, form *forms.EnableScanForm) (*models.PolicyRel, e.Error) {
	c.AddLogField("action", fmt.Sprintf("enable scan %s %s", form.Scope, form.Id))

	var (
		env *models.Env
		tpl *models.Template
		err e.Error
		rel *models.PolicyRel
	)

	query := c.DB()

	if form.Scope == consts.ScopeEnv {
		env, err = services.GetEnvById(query, form.Id)
		if err != nil {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
	} else {
		tpl, err = services.GetTemplateById(query, form.Id)
		if err != nil {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		}
	}

	rel, err = services.GetPolicyRel(query, form.Id, form.Scope)
	if err != nil && err.Code() != e.PolicyRelNotExist {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	// 添加启用关联
	attrs := models.Attrs{}
	if form.Enabled {
		if rel != nil {
			return rel, nil
		}

		if form.Scope == consts.ScopeEnv {
			rel = &models.PolicyRel{
				OrgId:     env.OrgId,
				ProjectId: env.ProjectId,
				EnvId:     env.Id,
			}
		} else {
			rel = &models.PolicyRel{
				OrgId:   tpl.OrgId,
				TplId:   tpl.Id,
			}
		}

		if _, err := services.CreatePolicyRel(query, rel); err != nil {
			return nil, e.New(err.Code(), err, http.StatusInternalServerError)
		}
		if form.Scope == consts.ScopeEnv {
			attrs["policyEnable"] = true
			if _, err := services.UpdateEnv(query, env.Id, attrs); err != nil {
				return nil, err
			}
		} else {
			attrs["policyEnable"] = true
			if _, err := services.UpdateTemplate(query, tpl.Id, attrs); err != nil {
				return nil, err
			}
		}

		return rel, nil
	} else {
		// 删除启用扫描关联
		if rel == nil {
			return nil, nil
		}

		if err := services.DeletePolicyEnabledRel(query, form.Id, form.Scope); err != nil {
			if err.Code() == e.PolicyRelNotExist {
				return nil, nil
			}
			return nil, e.New(err.Code(), err, http.StatusInternalServerError)
		}
		if form.Scope == consts.ScopeEnv {
			attrs["policyEnable"] = false
			if _, err := services.UpdateEnv(query, env.Id, attrs); err != nil {
				return nil, err
			}
		} else {
			attrs["policyEnable"] = false
			if _, err := services.UpdateTemplate(query, tpl.Id, attrs); err != nil {
				return nil, err
			}
		}
		
		return nil, nil
	}
}
