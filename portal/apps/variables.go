// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/desensitize"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"fmt"
	"net/http"
	"sort"
)

func BatchUpdate(c *ctx.ServiceContext, form *forms.BatchUpdateVariableForm) (interface{}, e.Error) {
	tx := c.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	err := services.OperationVariables(tx, c.OrgId, c.ProjectId, form.TplId, form.EnvId, form.Variables, form.DeleteVariablesId)
	if err != nil {
		c.Logger().Errorf("error creating variable, err %s", err)
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return nil, nil
}

func UpdateObjectVars(c *ctx.ServiceContext, form *forms.UpdateObjectVarsForm) (interface{}, e.Error) {
	var (
		result interface{}
		err    e.Error
	)
	_ = c.DB().Transaction(func(tx *db.Session) error {
		result, err = updateObjectVars(c, tx, form)
		return err
	})
	return result, err
}

func getObjectVars(tx *db.Session, form *forms.UpdateObjectVarsForm, orgId, projectId models.Id) ([]models.Variable, e.Error) {
	vars := make([]models.Variable, 0, len(form.Variables))
	for _, v := range form.Variables {
		if v.Scope != form.Scope {
			return nil, e.New(e.VariableScopeConflict, http.StatusBadRequest)
		}
		if v.Name == "" {
			return nil, e.New(e.EmptyVarName, http.StatusBadRequest)
		}

		// 目前前端是不允许变量值为空的，但第三方系统依然可能传入值为空的变量, 这里我们直接忽略掉
		// (sensitive 变量传入空值表示不修改其值)
		//if !v.Sensitive && v.Value == "" {
		//	continue
		//}

		modelVar := models.Variable{
			VariableBody: models.VariableBody{
				Scope:       v.Scope,
				Type:        v.Type,
				Name:        v.Name,
				Value:       v.Value,
				Sensitive:   v.Sensitive,
				Description: v.Description,
				Options:     v.Options,
			},
		}
		modelVar.Id = v.Id

		var (
			scope    = form.Scope
			objectId = form.ObjectId
		)

		switch scope {
		case consts.ScopeOrg:
			modelVar.OrgId = orgId
		case consts.ScopeProject:
			modelVar.OrgId = orgId
			modelVar.ProjectId = projectId
		case consts.ScopeTemplate:
			modelVar.OrgId = orgId
			modelVar.TplId = objectId
		case consts.ScopeEnv:
			modelVar.OrgId = orgId
			modelVar.ProjectId = projectId
			if env, er := services.GetEnvById(tx, objectId); er != nil {
				return nil, er
			} else {
				modelVar.TplId = env.TplId
			}
			modelVar.EnvId = objectId
		}
		vars = append(vars, modelVar)
	}

	return vars, nil
}

func updateObjectVars(c *ctx.ServiceContext, tx *db.Session, form *forms.UpdateObjectVarsForm) (interface{}, e.Error) {
	var (
		orgId     = c.OrgId
		projectId = c.ProjectId
		scope     = form.Scope
		objectId  = form.ObjectId
	)

	switch scope {
	case consts.ScopeOrg:
		if objectId != orgId {
			return nil, e.New(e.BadOrgId)
		}
	case consts.ScopeProject:
		if objectId != projectId {
			return nil, e.New(e.BadProjectId)
		}
	}

	vars, err := getObjectVars(tx, form, orgId, projectId)
	if err != nil {
		return nil, err
	}

	tx = services.QueryWithOrgId(tx, c.OrgId)
	retVars, err := services.UpdateObjectVars(tx, scope, objectId, vars)
	if err != nil {
		c.Logger().Warnf("update object %s(%s) vars error: %v", form.Scope, form.ObjectId, err)
		return nil, e.AutoNew(err, e.InternalError)
	}
	return services.VarsDesensitization(retVars), nil
}

func SearchVariable(c *ctx.ServiceContext, form *forms.SearchVariableForm) (interface{}, e.Error) {
	variableM, err, scopes := services.GetValidVariables(c.DB(), form.Scope, c.OrgId, c.ProjectId, form.TplId, form.EnvId, false)
	if err != nil {
		return nil, err
	}

	rs := make([]resps.VariableResp, 0)
	for _, variable := range variableM {
		vr := resps.VariableResp{
			Variable:   desensitize.NewVariable(variable),
			Overwrites: nil,
		}
		// 获取上一级被覆盖的变量
		varParent := services.GetVarParentParams{
			OrgId:        c.OrgId,
			ProjectId:    c.ProjectId,
			TplId:        form.TplId,
			Name:         variable.Name,
			Scope:        variable.Scope,
			VariableType: variable.Type,
		}

		if variable.Scope == form.Scope {
			isExists, overwrites := services.GetVariableParent(c.DB(), scopes, varParent)
			if isExists {
				vr.Overwrites = desensitize.NewVariablePtr(&overwrites)
			}
		}

		rs = append(rs, vr)
	}
	sort.Sort(resps.NewVariable(rs))

	return rs, nil
}

func SearchSampleVariable(c *ctx.ServiceContext, form *forms.SearchVariableForm) (interface{}, e.Error) {
	newRs := make([]models.VariableBody, 0)
	rs, err := SearchVariable(c, form)
	if err != nil {
		return nil, err
	}
	if rs != nil {
		for _, v := range rs.([]resps.VariableResp) {
			if v.Type == consts.VarTypeTerraform {
				newRs = append(newRs, models.VariableBody{
					Scope:       v.Scope,
					Type:        v.Type,
					Name:        fmt.Sprintf("TF_VAR_%s", v.Name),
					Value:       v.Value,
					Sensitive:   false,
					Description: v.Description,
				})
				continue
			}
			newRs = append(newRs, v.VariableBody)
		}
	}

	return newRs, nil
}
