// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"fmt"
	"net/http"
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
	if form.Provider != "alicloud" && form.CostCounted {
		return nil, e.New(e.BadParam, fmt.Errorf("cost statistics can only be enabled if the provider is alicloud"), http.StatusBadRequest)
	}
	session := c.DB()

	vb := make([]models.VarGroupVariable, 0)
	for index, v := range form.Variables {
		if v.Sensitive && v.Value != "" {
			value, err := utils.EncryptSecretVar(v.Value)
			if err != nil {
				return nil, e.AutoNew(err, e.EncryptError)
			}
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
		Name:        form.Name,
		Type:        form.Type,
		OrgId:       c.OrgId,
		CreatorId:   c.UserId,
		Variables:   vb,
		CostCounted: form.CostCounted,
		Provider:    form.Provider,
	})
	if err != nil {
		return nil, err
	}
	for _, projectId := range form.ProjectIds {
		err := services.CreateVariableGroupProjectRel(session, models.VariableGroupProjectRel{
			VarGroupId: vg.Id,
			ProjectId:  projectId,
		})
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	createVariableGroupResp := resps.CreateVariableGroupResp{
		VariableGroup: vg,
		ProjectIds:    form.ProjectIds,
	}

	return createVariableGroupResp, nil
}

func SearchVariableGroup(c *ctx.ServiceContext, form *forms.SearchVariableGroupForm) (interface{}, e.Error) {
	query := services.SearchVariableGroup(c.DB(), c.OrgId, c.ProjectId, form.Q)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	resp := make([]resps.SearchVariableGroupRespTemp, 0)

	if err := p.Scan(&resp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	resultTemp := make(map[models.Id][]string)

	for _, v := range resp {
		for index, variable := range v.Variables {
			if variable.Sensitive {
				v.Variables[index].Value = ""
			}
		}
		resultTemp[v.Id] = append(resultTemp[v.Id], v.ProjectName)
	}
	resArr := make([]resps.SearchVariableGroupRespTemp, 0)
	tmpMap := make(map[models.Id]interface{})
	for _, val := range resp {
		if _, ok := tmpMap[val.Id]; !ok {
			resArr = append(resArr, val)
			tmpMap[val.Id] = nil
		}
	}
	result := make([]resps.SearchVariableGroupResp, 0)
	for _, v := range resArr {
		restemp := resps.SearchVariableGroupResp{
			SearchVariableGroupRespTemp: v,
			ProjectNames:                resultTemp[v.Id],
		}
		result = append(result, restemp)
	}

	return page.PageResp{
		Total:    int64(len(result)),
		PageSize: p.Size,
		List:     result,
	}, nil
}

func getVarGroupVariables(variables []models.VarGroupVariable, vgVarsMap map[string]models.VarGroupVariable) ([]models.VarGroupVariable, e.Error) {
	vb := make([]models.VarGroupVariable, 0)
	for _, v := range variables {
		if v.Sensitive {
			if v.Value == "" {
				// 传空值时表示不修改，我们赋值为 db 中己保存的值(如果存在)
				v.Value = vgVarsMap[v.Name].Value
			} else {
				var err error
				v.Value, err = utils.EncryptSecretVar(v.Value)
				if err != nil {
					return nil, e.AutoNew(err, e.EncryptError)
				}
			}
		}
		vb = append(vb, models.VarGroupVariable{
			Id:          v.Id,
			Name:        v.Name,
			Value:       v.Value,
			Sensitive:   v.Sensitive,
			Description: v.Description,
		})
	}
	return vb, nil
}

func UpdateVariableGroup(c *ctx.ServiceContext, form *forms.UpdateVariableGroupForm) (interface{}, e.Error) {
	if form.Provider != "alicloud" && form.CostCounted {
		return nil, e.New(e.BadParam, fmt.Errorf("cost statistics can only be enabled if the provider is alicloud"), http.StatusBadRequest)
	}
	session := c.DB()
	attrs := models.Attrs{}

	// 修改变量组
	vg, er := services.GetVariableGroupById(session, form.Id)
	if er != nil {
		return nil, e.AutoNew(er, e.DBError)
	}

	vgVarsMap := make(map[string]models.VarGroupVariable)
	for _, v := range vg.Variables {
		vgVarsMap[v.Name] = v
	}

	if form.HasKey("variables") {
		vb, err := getVarGroupVariables(form.Variables, vgVarsMap)
		if err != nil {
			return nil, err
		}
		b, _ := models.VarGroupVariables(vb).Value()
		attrs["variables"] = b
	}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}
	if form.HasKey("provider") {
		attrs["provider"] = form.Provider
	}
	if form.HasKey("costCounted") {
		attrs["costCounted"] = form.CostCounted
	}
	err := session.Transaction(func(tx *db.Session) error {
		if err := services.UpdateVariableGroup(tx, form.Id, attrs); err != nil {
			return err
		}

		if form.HasKey("projectIds") {
			queryVgpRels, err := services.SearchVariableGroupProjectRelByVgpId(tx, form.Id)
			if err != nil {
				return err
			}
			delVgpRelsId, addVgpRelsId := services.GetDelVgpRelsIdAndAddVgpRelsId(queryVgpRels, form.ProjectIds)
			if err := services.UpdateVariableGroupProjectRel(tx, form.Id, delVgpRelsId, addVgpRelsId); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return nil, nil
}

func DeleteVariableGroup(c *ctx.ServiceContext, form *forms.DeleteVariableGroupForm) (interface{}, e.Error) {
	err := c.DB().Transaction(func(tx *db.Session) error {
		if err := services.DeleteVariableGroup(tx, form.Id); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return nil, nil
}

func DetailVariableGroup(c *ctx.ServiceContext, form *forms.DetailVariableGroupForm) (interface{}, e.Error) {
	vg := make([]resps.DetailVariableGroupRespTemp, 0)
	vgQuery := services.DetailVariableGroup(c.DB(), form.Id, c.OrgId)
	if err := vgQuery.Scan(&vg); err != nil {
		return nil, e.New(e.DBError, err)
	}
	projectNames := []string{}
	projectIds := []string{}
	for _, v := range vg {
		projectNames = append(projectNames, v.ProjectName)
		projectIds = append(projectIds, v.ProjectId)
	}
	variableResp := vg[0]
	for index, v := range variableResp.Variables {
		if v.Sensitive {
			variableResp.Variables[index].Value = ""
		}
	}
	result := resps.DetailVariableGroupResp{
		DetailVariableGroupRespTemp: variableResp,
		ProjectNames:                projectNames,
		ProjectIds:                  projectIds,
	}
	return result, nil
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
	return SetSensitiveNull(vgs), nil
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
		_ = tx.Rollback()
		return nil, err
	}

	if services.CheckVgRelationship(tx, form, c.OrgId) {
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

func SearchRelationshipAll(c *ctx.ServiceContext, form *forms.SearchRelationshipForm) (interface{}, e.Error) {
	// 继承逻辑 当前作用域下的变量组包含变量组中的变量时 进行覆盖
	// 查询作用域下的所有变量
	resp := make([]services.VarGroupRel, 0)
	vgs, err := services.SearchVariableGroupRel(c.DB(), map[string]models.Id{
		consts.ScopeEnv:      form.EnvId,
		consts.ScopeTemplate: form.TplId,
		consts.ScopeProject:  c.ProjectId,
		consts.ScopeOrg:      c.OrgId,
	}, form.ObjectType)
	if err != nil {
		return nil, err
	}

	for _, v := range vgs {
		resp = append(resp, v)
		if len(v.Overwrites) != 0 && v.ObjectType == form.ObjectType {
			resp = append(resp, v.Overwrites...)
		}
	}
	//按照id去重
	return DelDuplicate(SetSensitiveNull(resp)), nil
}

func DelDuplicate(arr []services.VarGroupRel) []services.VarGroupRel {
	rel := make([]services.VarGroupRel, 0)
	for _, v := range arr {
		isDuplicate := false
		for _, r := range rel {
			if v.Id == r.Id {
				isDuplicate = true
			}
		}
		if !isDuplicate {
			rel = append(rel, v)
		}

	}
	return rel
}

func SetSensitiveNull(vg []services.VarGroupRel) []services.VarGroupRel {
	for _, v := range vg {
		for index, variable := range v.Variables {
			if variable.Sensitive {
				v.Variables[index].Value = ""
			}
		}
	}
	return vg
}
