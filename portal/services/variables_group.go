package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"fmt"
)

func CreateVariableGroup(tx *db.Session, group models.VariableGroup) (models.VariableGroup, e.Error) {
	if group.Id == "" {
		group.Id = models.NewId("vg")
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
	if err := DeleteRelationship(tx, vgId); err != nil {
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
}

func SearchVariableGroupRel(dbSess *db.Session, objectAttr map[string]models.Id, object string) ([]VarGroupRel, e.Error) {
	scopes := make([]string, 0)
	switch object {
	case consts.ScopeEnv:
		scopes = consts.VariableGroupEnv
	case consts.ScopeTemplate:
		scopes = consts.VariableGroupTpl
	case consts.ScopeProject:
		scopes = consts.VariableGroupProject
	case consts.ScopeOrg:
		scopes = consts.VariableGroupOrg
	}
	// {objectType:{objectId:xxx}}
	rels := make(map[models.Id]VarGroupRel, 0)

	// 按照继承顺序一层一层查询对应的变量组数据
	for index, v := range scopes {
		addRels := make([]VarGroupRel, 0)
		// 查询当前作用域下的变量组信息
		vgs, err := GetVariableGroupByObject(dbSess, v, objectAttr[v])
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
		resp = append(resp, v)
	}

	return resp, nil
}

func GetVariableGroupByObject(dbSess *db.Session, objectType string, objectId models.Id) ([]VarGroupRel, e.Error) {
	vg := make([]VarGroupRel, 0)
	query := dbSess.Table(fmt.Sprintf("%s as rel", models.VariableGroupRel{}.TableName())).
		Where("rel.object_id = ?", objectId).
		Where("rel.object_type = ?", objectType)
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

func SearchRelationship(dbSess *db.Session, vgId, orgId models.Id) *db.Session {
	return dbSess
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

func DeleteRelationship(dbSess *db.Session, vgId models.Id) e.Error {
	if _, err := dbSess.Where("var_group_id = ?", vgId).
		Delete(&models.VariableGroupRel{}); err != nil {
		return e.New(e.DBError, err)
	}
	return nil
}
