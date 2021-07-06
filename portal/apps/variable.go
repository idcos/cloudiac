package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
)

func CreateVariable(c *ctx.ServiceCtx, form *forms.CreateVariableForm) (interface{}, e.Error) {
	return services.CreateVariable(c.DB(), models.Variable{
		BaseModel: models.BaseModel{},
		VariableBody: models.VariableBody{
			Scope:       form.Scope,
			Type:        form.Type,
			Name:        form.Name,
			Value:       form.Value,
			Sensitive:   form.Sensitive,
			Description: form.Description,
		},
	})
}

func SearchVariable(c *ctx.ServiceCtx, form *forms.SearchVariableForm) (interface{}, e.Error) {
	query := services.SearchVariable(c.DB(), form.Id)
	// 默认按创建时间逆序排序
	if form.SortField() == "" {
		query = query.Order("created_at DESC")
	}
	rs, err := getPage(query, form, &models.Variable{})
	if err != nil {
		c.Logger().Errorf("error get page, err %s", err)
	}
	return rs, err
}
