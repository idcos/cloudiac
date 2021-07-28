package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"sort"
)

func BatchUpdate(c *ctx.ServiceContext, form *forms.BatchUpdateVariableForm) (interface{}, e.Error) {
	tx := c.DB().Begin().Debug()
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

type newVariable []VariableResp

func (v newVariable) Len() int {
	return len(v)
}
func (v newVariable) Less(i, j int) bool {
	return v[i].Name < v[j].Name
}
func (v newVariable) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

type VariableResp struct {
	models.Variable
	Overwrites *models.Variable `json:"overwrites" form:"overwrites" ` //回滚参数，无需回滚是为空
}

func SearchVariable(c *ctx.ServiceContext, form *forms.SearchVariableForm) (interface{}, e.Error) {
	variableM, err, scopes := services.GetValidVariables(c.DB(), form.Scope, c.OrgId, c.ProjectId, form.TplId, form.EnvId, false)
	if err != nil {
		return nil, err
	}
	rs := make([]VariableResp, 0)
	for _, variable := range variableM {
		vr := VariableResp{
			Variable:   variable,
			Overwrites: nil,
		}
		// 获取上一级被覆盖的变量
		if variable.Scope == form.Scope {
			isExists, overwrites := services.GetVariableParent(c.DB(), variable.Name, variable.Scope, variable.Type, scopes, c.OrgId, c.ProjectId, form.TplId)
			if isExists {
				if overwrites.Sensitive {
					overwrites.Value = ""
				}
				vr.Overwrites = &overwrites
			}
		}

		rs = append(rs, vr)
	}
	sort.Sort(newVariable(rs))

	return rs, nil
}
