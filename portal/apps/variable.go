package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"net/http"
)

func CreateVariable(c *ctx.ServiceCtx, form *forms.CreateVariableForm) (interface{}, e.Error) {
	// todo 如何校验权限
	tx := c.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()
	// 新建variable
	variables, err := services.BindVariable(tx, c.OrgId, c.ProjectId, form.TplId, form.EnvId, form.Variables)
	if err != nil {
		c.Logger().Errorf("error creating variable, err %s", err)
		_ = tx.Rollback()
		return nil, err
	}

	// 修改variable
	for _, v := range variables {
		var (
			value string
		)

		if v.Sensitive && v.Value != "" {
			value, _ = utils.AesEncrypt(v.Value)
		}
		if v.Value == "" && v.Sensitive {
			value = v.Value
		}
		attrs := map[string]interface{}{
			"name":        v.Name,
			"value":       value,
			"sensitive":   v.Sensitive,
			"description": v.Description,
		}
		err := services.UpdateVariable(tx, v.Id, attrs)
		if err != nil && err.Code() == e.VariableAliasDuplicate {
			return nil, e.New(err.Code(), err, http.StatusBadRequest)
		} else if err != nil {
			c.Logger().Errorf("error updating variable, err %s", err)
			return nil, err
		}
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
	query := services.SearchVariable(c.DB(), c.OrgId, c.ProjectId, form.TplId, form.EnvId, form.Scope)
	rs, err := getPage(query, form, &models.Variable{})
	if err != nil {
		c.Logger().Errorf("error get page, err %s", err)
	}
	return rs, err
}
