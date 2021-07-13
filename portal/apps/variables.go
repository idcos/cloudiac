package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
)

func CreateVariable(c *ctx.ServiceCtx, form *forms.CreateVariableForm) (interface{}, e.Error) {
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
	variableM, err := services.GetValidVariables(c.DB(), form.Scope, c.OrgId, c.ProjectId, form.TplId, form.EnvId)
	if err != nil {
		return nil, err
	}
	rs := make([]models.Variable, 0)
	for _, variable := range variableM {
		rs = append(rs, variable)
	}

	return rs, nil
}
