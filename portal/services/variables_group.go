package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/utils/logs"
	"fmt"
)

func CreateVariableGroup(tx *db.Session, group models.VariableGroup) (models.VariableGroup, e.Error) {
	if group.Id == "" {
		group.Id = group.NewId()
	}
	if err := models.Create(tx, &group); err != nil {
		if e.IsDuplicate(err) {
			return group, e.New(e.VariableGroupAlreadyExist, err)
		}
		return group, e.AutoNew(err, e.DBError)
	}
	return group, nil
}

func SearchVariableGroup(dbSess *db.Session, orgId models.Id, q string) *db.Session {
	query := dbSess.Model(models.VariableGroup{}).Where("iac_variable_group.org_id = ?", orgId)
	if q != "" {
		query = query.WhereLike("iac_variable_group.name", q)
	}
	return query.Joins("left join iac_user as u on u.id = iac_variable_group.creator_id").
		LazySelectAppend("iac_variable_group.*").
		LazySelectAppend("u.name as creator")

}

func UpdateVariableGroup(tx *db.Session, id models.Id, attrs models.Attrs) e.Error {
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.VariableGroup{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return e.New(e.VariableGroupAliasDuplicate)
		} else if e.IsRecordNotFound(err) {
			return e.New(e.VariableGroupNotExist)
		}
		return e.New(e.DBError, fmt.Errorf("update variable group error: %v", err))
	}
	return nil
}

func DeleteVariableGroup(tx *db.Session, vgId models.Id) e.Error {
	//删除变量组
	if _, err := tx.Where("id = ?", vgId).Delete(&models.VariableGroup{}); err != nil {
		return e.New(e.DBError, err)
	}

	//删除变量组与实例之间的关系
	if err := DeleteRelationship(tx, []models.Id{vgId}); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

func DetailVariableGroup(dbSess *db.Session, vgId, orgId models.Id) *db.Session {
	return dbSess.Model(&models.VariableGroup{}).
		Where("id = ?", vgId).
		Where("org_id = ?", orgId)
}

type VarGroupRel struct {
	models.VariableGroupRel
	models.VariableGroup
	Overwrites []VarGroupRel `json:"overwrites" form:"overwrites" gorm:"-"` //回滚参数，无需回滚是为空
}

func SearchVariableGroupRel(dbSess *db.Session, objectAttr map[string]models.Id, object string) ([]VarGroupRel, e.Error) {
	scopes := make([]string, 0)
	switch object {
	case consts.ScopeEnv:
		scopes = consts.VariableGroupEnv
		// 环境id为空时，只有新部署环境的场景，这里只要查询org/tpl/project作用域的变量组即可
		if objectAttr[consts.ScopeEnv] == "" {
			scopes = []string{consts.ScopeOrg, consts.ScopeProject, consts.ScopeTemplate}
		}
	case consts.ScopeTemplate:
		scopes = consts.VariableGroupTpl
	case consts.ScopeProject:
		scopes = consts.VariableGroupProject
	case consts.ScopeOrg:
		scopes = consts.VariableGroupOrg
	}
	// {objectType:{objectId:xxx}}
	rels := make(map[models.Id]VarGroupRel, 0)

	coverRels := make(map[models.Id][]VarGroupRel)

	// 按照继承顺序一层一层查询对应的变量组数据
	for index, v := range scopes {
		addRels := make([]VarGroupRel, 0)
		// 查询当前作用域下的变量组信息
		vgs, err := GetVariableGroupByObject(dbSess, v, objectAttr[v], objectAttr[consts.ScopeOrg])
		if err != nil {
			continue
		}

		for _, vg := range vgs {
			// 如果进行初始化则直接写入Map
			if _, ok := rels[vg.VarGroupId]; !ok && index == 0 {
				rels[vg.VarGroupId] = vg
				continue
			}
			// 比较是否有需要覆盖的变量
			for k, v := range rels {
				if MatchVarGroup(v, vg) {
					// 需要覆盖则删除上一级的变量组
					delete(rels, k)
					// 记录覆盖的变量
					if _, ok := coverRels[vg.VarGroupId]; !ok {
						coverRels[vg.VarGroupId] = []VarGroupRel{
							v,
						}
						continue
					}
					coverRels[vg.VarGroupId] = append(coverRels[vg.VarGroupId], v)
				}

			}
			//临时存储需要添加的变量组,避免重复相同层级的变量进行比较
			addRels = append(addRels, vg)
		}

		//进行批量添加
		for _, rel := range addRels {
			rels[rel.VarGroupId] = rel
		}
	}

	// 整理数据并返回
	resp := make([]VarGroupRel, 0)
	for _, v := range rels {
		v.Overwrites = coverRels[v.VarGroupId]
		resp = append(resp, v)
	}

	return resp, nil
}

func GetVariableGroupByObject(dbSess *db.Session, objectType string, objectId, orgId models.Id) ([]VarGroupRel, e.Error) {
	vg := make([]VarGroupRel, 0)
	query := dbSess.Table(fmt.Sprintf("%s as rel", models.VariableGroupRel{}.TableName())).
		Where("vg.org_id = ?", orgId)

	if objectType != "" {
		query = query.Where("rel.object_type = ?", objectType)
	}
	if objectId != "" {
		query = query.Where("rel.object_id = ?", objectId)
	}
	query = query.LazySelectAppend("rel.*")
	query = query.LazySelectAppend("vg.*")
	if err := query.Joins("left join iac_variable_group as vg on vg.id = rel.var_group_id").Find(&vg); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return vg, nil
}

//MatchVarGroup 有相同的name 返回true 没有返回false
func MatchVarGroup(oldVg, newVg VarGroupRel) bool {
	for _, old := range oldVg.Variables {
		for _, v := range newVg.Variables {
			if old.Name == v.Name {
				return true
			}
		}
	}
	return false

}

func CreateRelationship(dbSess *db.Session, rels []models.VariableGroupRel) e.Error {
	if len(rels) == 0 {
		return nil
	}
	if err := models.CreateBatch(dbSess, &rels); err != nil {
		if e.IsDuplicate(err) {
			return e.New(e.VariableGroupAlreadyExist, err)
		}
		return e.AutoNew(err, e.DBError)
	}
	return nil
}

func CheckVgRelationship(tx *db.Session, form *forms.BatchUpdateRelationshipForm, orgId models.Id) bool {
	// 查询当前作用域下绑定的变量组
	bindVgs, err := GetVariableGroupByObject(tx, form.ObjectType, form.ObjectId, orgId)
	if err != nil {
		logs.Get().Errorf("func GetVariableGroupByObject err: %v", err)
		return false
	}
	// 查询将要绑定的变量组
	notBindVgs, err := GetVariableGroupListByIds(tx, form.VarGroupIds)
	if err != nil {
		logs.Get().Errorf("func GetVariableGroupListByIds err: %v", err)
		return false
	}
	// 比较变量组下变量是否与其他变量组下变量存在冲突
	// 利用map将当前作用域下绑定的变量组变量进行整理
	variables := make(map[string]interface{})
	for _, bindVg := range bindVgs {
		for _, vg := range bindVg.Variables {
			variables[vg.Name] = vg.Value
		}
	}

	for _, notBindVg := range notBindVgs {
		for _, vg := range notBindVg.Variables {
			// 校验新绑定的变量组变量是否冲突
			if _, ok := variables[vg.Name]; ok {
				return true
			}
		}
	}
	return false
}

func GetVariableGroupById(dbSess *db.Session, id models.Id) (models.VariableGroup, e.Error) {
	vg := models.VariableGroup{}
	if err := dbSess.Where("id = ?", id).First(&vg); err != nil {
		return vg, e.New(e.DBError, err)
	}
	return vg, nil
}

func GetVariableGroupListByIds(dbSess *db.Session, ids []models.Id) ([]models.VariableGroup, e.Error) {
	vgs := make([]models.VariableGroup, 0)
	if len(ids) == 0 {
		return nil, e.New(e.BadParam, fmt.Errorf("id list is null"))
	}
	if err := dbSess.Where("id in (?)", ids).Find(&vgs); err != nil {
		return vgs, e.New(e.DBError, err)
	}
	return vgs, nil
}

func DeleteRelationship(dbSess *db.Session, vgId []models.Id) e.Error {
	if len(vgId) == 0 {
		return nil
	}
	if _, err := dbSess.Where("var_group_id in (?)", vgId).
		Delete(&models.VariableGroupRel{}); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}

func GetVariableGroupVar(vgs []VarGroupRel, vars map[string]models.Variable) map[string]models.Variable {
	variableM := make(map[string]models.Variable)
	//newVariableM := make(map[string]models.Variable)
	for _, v := range vgs {
		for _, variable := range v.Variables {
			variableM[fmt.Sprintf("%s%s", variable.Name, v.Type)] = models.Variable{
				VariableBody: models.VariableBody{
					Scope:       v.ObjectType,
					Type:        v.Type,
					Name:        variable.Name,
					Value:       variable.Value,
					Sensitive:   variable.Sensitive,
					Description: variable.Description,
				},
			}
		}
	}
	// 将标准变量覆盖
	for k, v := range vars {
		variableM[k] = v
	}
	return variableM
}

func BatchUpdateRelationship(tx *db.Session, vgIds, delVgIds []models.Id, objectType, objectId string) e.Error {
	rel := make([]models.VariableGroupRel, 0)
	if err := DeleteRelationship(tx, delVgIds); err != nil {
		return err
	}

	for _, v := range vgIds {
		rel = append(rel, models.VariableGroupRel{
			VarGroupId: v,
			ObjectType: objectType,
			ObjectId:   models.Id(objectId),
		})
	}

	if err := CreateRelationship(tx, rel); err != nil {
		return err
	}
	return nil
}

// GetValidVarsAndVgVars 获取变量及变量组变量
func GetValidVarsAndVgVars(tx *db.Session, orgId, projectId, tplId, envId models.Id) ([]models.VariableBody, error) {
	vars, err, _ := GetValidVariables(tx, consts.ScopeEnv, orgId, projectId, tplId, envId, true)
	if err != nil {
		return nil, fmt.Errorf("get vairables error: %v", err)
	}

	// 将变量组变量与普通变量进行合并，优先级: 普通变量 > 变量组变量
	// 查询实例关联的变量组
	varGroup, err := SearchVariableGroupRel(tx, map[string]models.Id{
		consts.ScopeEnv:      envId,
		consts.ScopeTemplate: tplId,
		consts.ScopeProject:  projectId,
		consts.ScopeOrg:      orgId,
	}, consts.ScopeEnv)

	if err != nil {
		return nil, fmt.Errorf("get vairable group var error: %v", err)
	}
	return GetVariableBody(GetVariableGroupVar(varGroup, vars)), nil
}

// 查询指定模板直接关联的变量组
func FindTplsRelVarGroup(query *db.Session, tplIds []models.Id) ([]models.VariableGroup, error) {
	vgs := make([]models.VariableGroup, 0)
	err := query.Table("iac_variable_group AS vg").
		Joins("JOIN iac_variable_group_rel AS vgr ON vgr.var_group_id = vg.id").
		Where("vgr.object_type = ? AND vgr.object_id IN (?)", consts.ScopeTemplate, tplIds).
		Select("vg.*").Find(&vgs)
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return vgs, nil
}

func QueryVarGroup(sess *db.Session) *db.Session {
	return sess.Model(&models.VariableGroup{})
}

func FindTemplateVgIds(sess *db.Session, tplId models.Id) ([]models.Id, e.Error) {
	ids := make([]models.Id, 0)
	err := sess.Model(&models.VariableGroupRel{}).
		Where("object_type = ? AND object_id = ?", consts.ScopeTemplate, tplId).Pluck("var_group_id", &ids)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	return ids, nil
}

func DeleteVarGroupRel(sess *db.Session, objectType string, objectId models.Id) e.Error {
	_, err := sess.
		Where("object_type = ? AND object_id = ?", objectType, objectId).
		Delete(&models.VariableGroupRel{})
	if err != nil {
		return e.AutoNew(err, e.DBError)
	}
	return nil
}
