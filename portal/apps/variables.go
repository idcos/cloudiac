package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"sort"
)

func BatchUpdate(c *ctx.ServiceCtx, form *forms.BatchUpdateVariableForm) (interface{}, e.Error) {
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

	if err := services.DeleteVariables(tx, form.DeleteVariablesId); err != nil {
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

type newVariable []models.Variable

func (v newVariable) Len() int {
	return len(v)
}
func (v newVariable) Less(i, j int) bool {
	return v[i].CreatedAt.Unix() < v[j].CreatedAt.Unix()
}
func (v newVariable) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
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
	sort.Sort(newVariable(rs))

	return rs, nil
}
