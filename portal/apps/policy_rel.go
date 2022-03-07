// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils/logs"
	"fmt"
	"net/http"
)

func UpdatePolicyRelNew(c *ctx.ServiceContext, form *forms.UpdatePolicyRelForm) ([]*models.PolicyRel, e.Error) {
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	result, err := services.UpdatePolicyRel(tx, form)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		logs.Get().Errorf("error commit policy relations, err %s", err)
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}
	return result, nil
}

// EnablePolicyScanRel 启用环境/云模板扫描
func EnablePolicyScanRel(c *ctx.ServiceContext, form *forms.EnableScanForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("enable scan %s %s", form.Scope, form.Id))

	var (
		env *models.Env
		tpl *models.Template
		err e.Error
	)

	query := services.QueryWithOrgId(c.DB(), c.OrgId)

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

	// 添加启用关联
	attrs := models.Attrs{}
	if form.Enabled {
		attrs["policyEnable"] = true
		if form.Scope == consts.ScopeEnv {
			if _, err := services.UpdateEnv(query, env.Id, attrs); err != nil {
				return nil, err
			}
		} else {
			if _, err := services.UpdateTemplate(query, tpl.Id, attrs); err != nil {
				return nil, err
			}
		}
	} else {
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
	}
	return nil, nil
}
