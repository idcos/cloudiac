package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
)

func CreateVariable(c *ctx.ServiceCtx, form *forms.CreateVariableForm) (interface{}, e.Error) {
	// todo 如何校验权限
	tx := c.DB().Begin().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	err := services.OperationVariables(tx, c.OrgId, c.ProjectId, form.TplId, form.EnvId, form.Variables)
	if err != nil {
		c.Logger().Errorf("error creating variable, err %s", err)
		_ = tx.Rollback()
		return nil, err
	}

	if err := services.DeleteVariables(tx, form.DeleteVariables); err != nil {
		c.Logger().Errorf("error delete variable, err %s", err)
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return nil, nil
}

func SearchVariable(c *ctx.ServiceCtx, form *forms.SearchVariableForm) (interface{}, e.Error) {
	var (
		scopeEnv     = []string{consts.ScopeEnv, consts.ScopeTemplate, consts.ScopeProject, consts.ScopeOrg}
		scopeTpl     = []string{consts.ScopeTemplate, consts.ScopeProject, consts.ScopeOrg}
		scopeProject = []string{consts.ScopeProject, consts.ScopeOrg}
		scopeOrg     = []string{consts.ScopeOrg}
	)
	scopes := make([]string, 0)
	switch form.Scope {
	case "env":
		scopes = scopeEnv
	case "tpl":
		scopes = scopeTpl
	case "project":
		scopes = scopeProject
	case "org":
		scopes = scopeOrg
	}

	variables, err := services.SearchVariable(c.DB().Debug(), c.OrgId, c.ProjectId, form.TplId, form.EnvId, form.Scope)
	if err != nil {
		c.Logger().Errorf("error get variables, err %s", err)
		return nil, err
	}
	variableM := make(map[string]models.Variable, 0)
	for _, v := range variables {
		if utils.ArrayIsHasSuffix(scopes, v.Scope) {
			if v.EnvId != "" && v.EnvId == form.EnvId {
				variableM[v.Name] = v
				continue
			}

			if v.TplId != "" && v.TplId == form.TplId {
				variableM[v.Name] = v
				continue
			}

			if v.ProjectId != "" && v.ProjectId == c.ProjectId {
				variableM[v.Name] = v
				continue
			}

			variableM[v.Name] = v
		}
	}
	rs := make([]models.Variable, 0)
	for _, variable := range variableM {
		rs = append(rs, variable)
	}

	return rs, nil
}
