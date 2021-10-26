package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"fmt"
)

type CreateVariableGroupForm struct {
	Name              string                    `json:"name" form:"name"`
	Type              string                    `json:"type" form:"type"`
	VarGroupVariables []VarGroupVariablesCreate `json:"varGroupVariables" form:"varGroupVariables" `
}

type VarGroupVariablesCreate struct {
	Name        string `json:"name" form:"name" `
	Value       string `json:"value" form:"value" `
	Sensitive   bool   `json:"sensitive" form:"sensitive" `
	Description string `json:"description" form:"description" `
}

func CreateVariableGroup(c *ctx.ServiceContext, form *forms.CreateVariableGroupForm) (interface{}, e.Error) {
	session := c.DB()

	vb := make([]models.VarGroupVariable, 0)
	for index, v := range form.Variables {
		if v.Sensitive {
			value, _ := utils.AesEncrypt(v.Value)
			form.Variables[index].Value = value
		}
		vb = append(vb, models.VarGroupVariable{
			Id:          form.Variables[index].Id,
			Name:        form.Variables[index].Name,
			Value:       form.Variables[index].Value,
			Sensitive:   form.Variables[index].Sensitive,
			Description: form.Variables[index].Description,
		})
	}
	// 创建变量组
	vg, err := services.CreateVariableGroup(session, models.VariableGroup{
		Name:      form.Name,
		Type:      form.Type,
		OrgId:     c.OrgId,
		CreatorId: c.UserId,
		Variables: models.VarGroupVariables(vb),
	})
	if err != nil {
		return nil, err
	}

	return vg, nil
}

type SearchVariableGroupResp struct {
	models.VariableGroup
	Creator string `json:"creator" form:"creator" `
}

func SearchVariableGroup(c *ctx.ServiceContext, form *forms.SearchVariableGroupForm) (interface{}, e.Error) {
	query := services.SearchVariableGroup(c.DB(), c.OrgId, form.Q)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	resp := make([]SearchVariableGroupResp, 0)
	if err := p.Scan(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     resp,
	}, nil
}

func UpdateVariableGroup(c *ctx.ServiceContext, form *forms.UpdateVariableGroupForm) (interface{}, e.Error) {
	session := c.DB()
	attrs := models.Attrs{}

	// 修改变量组
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("variables") {
		vb := make([]models.VarGroupVariable, 0)
		for index, v := range form.Variables {
			if v.Sensitive {
				value, _ := utils.AesEncrypt(v.Value)
				form.Variables[index].Value = value
			}
			vb = append(vb, models.VarGroupVariable{
				Id:          form.Variables[index].Id,
				Name:        form.Variables[index].Name,
				Value:       form.Variables[index].Value,
				Sensitive:   form.Variables[index].Sensitive,
				Description: form.Variables[index].Description,
			})
		}
		b, _ := models.VarGroupVariables(vb).Value()
		attrs["variables"] = b
	}

	if err := services.UpdateVariableGroup(session, form.Id, attrs); err != nil {
		return nil, err
	}

	return nil, nil
}

func DeleteVariableGroup(c *ctx.ServiceContext, form *forms.DeleteVariableGroupForm) (interface{}, e.Error) {
	session := c.DB()
	if err := services.DeleteVariableGroup(session, form.Id); err != nil {
		return nil, err
	}
	return nil, nil
}

func DetailVariableGroup(c *ctx.ServiceContext, form *forms.DetailVariableGroupForm) (interface{}, e.Error) {
	vg := models.VariableGroup{}
	vgQuery := services.DetailVariableGroup(c.DB(), form.Id, c.OrgId)
	if err := vgQuery.First(&vg); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return vg, nil
}

func SearchRelationship(c *ctx.ServiceContext, form *forms.SearchRelationshipForm) (interface{}, e.Error) {
	// 继承逻辑 当前作用域下的变量组包含变量组中的变量时 进行覆盖
	// 查询作用域下的所有变量
	vgs, err := services.SearchVariableGroupRel(c.DB(), map[string]models.Id{
		consts.ScopeEnv:      form.EnvId,
		consts.ScopeTemplate: form.TplId,
		consts.ScopeProject:  c.ProjectId,
		consts.ScopeOrg:      c.OrgId,
	}, form.ObjectType)
	if err != nil {
		return nil, err
	}
	return vgs, nil
}

func BatchUpdateRelationship(c *ctx.ServiceContext, form *forms.BatchUpdateRelationshipForm) (interface{}, e.Error) {
	rel := make([]models.VariableGroupRel, 0)
	// 校验变量组在同级是否有相同key的变量
	tx := c.Tx()

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if err := services.DeleteRelationship(tx, form.DelVarGroupIds); err != nil {
		return nil, err
	}
	fmt.Println(form.DelVarGroupIds)
	if services.CheckVgRelationship(tx, form) {
		_ = tx.Rollback()
		return nil, e.New(e.VariableAlreadyExists, fmt.Errorf("the variables under the variable group are repeated"))
	}

	for _, v := range form.VarGroupIds {
		rel = append(rel, models.VariableGroupRel{
			VarGroupId: v,
			ObjectType: form.ObjectType,
			ObjectId:   form.ObjectId,
		})
	}
	if err := services.CreateRelationship(tx, rel); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, err)
	}

	return nil, nil
}

func DeleteRelationship(c *ctx.ServiceContext, form *forms.DeleteRelationshipForm) (interface{}, e.Error) {
	//if err := services.DeleteRelationship(c.DB(), form.Id); err != nil {
	//	return nil, err
	//}
	return nil, nil
}
