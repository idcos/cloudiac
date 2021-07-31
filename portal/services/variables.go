// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/common"
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/utils"
	"fmt"
	"net/http"
)

func CreateVariable(tx *db.Session, variable models.Variable) (*models.Variable, e.Error) {
	if variable.Id == "" {
		variable.Id = models.NewId("v")
	}
	if err := models.Create(tx, &variable); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.VariableAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &variable, nil
}

func SearchVariable(dbSess *db.Session, orgId models.Id) ([]models.Variable, e.Error) {
	variables := make([]models.Variable, 0)
	if err := dbSess.Model(models.Variable{}.TableName()).
		Where("org_id = ?", orgId).
		// 按照枚举值排序控制org类型在最上面
		Order("scope asc").
		Find(&variables); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return variables, nil
}

func SearchVariableByTemplateId(tx *db.Session, tplId models.Id) ([]models.Variable, e.Error) {
	variables := make([]models.Variable, 0)
	if err := tx.Where("tpl_id = ?", tplId).Find(&variables); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return variables, nil
}

func OperationVariables(tx *db.Session, orgId, projectId, tplId, envId models.Id,
	variables []forms.Variables, deleteVariablesId []string) e.Error {
	if err := DeleteVariables(tx, deleteVariablesId); err != nil {
		return err
	}

	bq := utils.NewBatchSQL(1024, "INSERT INTO", models.Variable{}.TableName(),
		"id", "scope", "type", "name", "value", "sensitive", "description", "org_id", "project_id", "tpl_id", "env_id")
	for _, v := range variables {
		attrs := map[string]interface{}{
			"name":        v.Name,
			"sensitive":   v.Sensitive,
			"description": v.Description,
		}
		var value string = v.Value
		// 需要加密，数据不为空
		if v.Sensitive && v.Value != "" {
			value, _ = utils.AesEncrypt(v.Value)
			attrs["value"] = value
		}

		// 不需要加密，数据不为空
		if v.Value != "" && !v.Sensitive {
			value = v.Value
			attrs["value"] = value
		}
		// 需要加密，数据为空 不做操作

		//id不为空修改变量，反之新建
		if v.Id != "" {
			err := UpdateVariable(tx, v.Id, attrs)
			if err != nil && err.Code() == e.VariableAliasDuplicate {
				return e.New(err.Code(), err, http.StatusBadRequest)
			} else if err != nil {
				return err
			}
			continue
		} else {
			vId := models.NewId("v")
			if err := bq.AddRow(vId, v.Scope, v.Type, v.Name, value, v.Sensitive, v.Description,
				orgId, projectId, tplId, envId); err != nil {
				return e.New(e.DBError, err)
			}
		}
	}
	if err := CreateVariables(tx, bq); err != nil {
		return err
	}
	return nil
}

func CreateVariables(tx *db.Session, bq *utils.BatchSQL) e.Error {
	for bq.HasNext() {
		sql, args := bq.Next()
		if _, err := tx.Exec(sql, args...); err != nil {
			return e.New(e.DBError, err)
		}
	}
	return nil
}

func UpdateVariable(tx *db.Session, variableId models.Id, attr map[string]interface{}) e.Error {
	if _, err := models.UpdateAttr(tx,
		models.Variable{BaseModel: models.BaseModel{Id: variableId}}, attr); err != nil {
		if e.IsDuplicate(err) {
			return e.New(e.VariableAliasDuplicate)
		}
		return e.New(e.DBError, fmt.Errorf("update variable error: %v", err))
	}
	return nil
}

func DeleteVariables(tx *db.Session, DeleteVariables []string) e.Error {
	if len(DeleteVariables) == 0 {
		return nil
	}
	if _, err := tx.Where("id in (?)", DeleteVariables).Delete(&models.Variable{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete variables error: %v", err))
	}
	return nil
}

func GetValidVariables(dbSess *db.Session, scope string, orgId, projectId, tplId, envId models.Id, keepSensitive bool) (map[string]models.Variable, e.Error, []string) {
	// 根据scope 构建变量应用范围
	scopes := make([]string, 0)
	switch scope {
	case consts.ScopeEnv:
		scopes = common.EnvScopeEnv
	case consts.ScopeTemplate:
		scopes = common.EnvScopeTpl
	case consts.ScopeProject:
		scopes = common.EnvScopeProject
	case consts.ScopeOrg:
		scopes = common.EnvScopeOrg
	}

	// 将组织下所有的变量查询，在代码处理变量的继承关系及是否要应用该变量
	variables, err := SearchVariable(dbSess, orgId)
	if err != nil {
		return nil, err, scopes
	}
	variableM := make(map[string]models.Variable, 0)
	for index, v := range variables {
		// 过滤掉变量一部分不需要应用的变量
		if utils.InArrayStr(scopes, v.Scope) {
			if v.Sensitive && !keepSensitive {
				variables[index].Value = ""
			}
			// 根据id（envId/tplId/projectId）来确认变量是否需要应用
			if v.EnvId != "" {
				if v.EnvId == envId {
					// 不同的变量类型也有可能出现相同的name
					variableM[fmt.Sprintf("%s%s", v.Name, v.Type)] = variables[index]
				}
				continue
			}

			if v.TplId != "" {
				if v.TplId == tplId {
					// 不同的变量类型也有可能出现相同的name
					variableM[fmt.Sprintf("%s%s", v.Name, v.Type)] = variables[index]
				}
				continue
			}

			if v.ProjectId != "" {
				if v.ProjectId == projectId {
					// 不同的变量类型也有可能出现相同的name
					variableM[fmt.Sprintf("%s%s", v.Name, v.Type)] = variables[index]
				}
				continue
			}

			if v.ProjectId == "" && v.TplId == "" && v.EnvId == "" {
				variableM[fmt.Sprintf("%s%s", v.Name, v.Type)] = variables[index]
			}

		}
	}

	return variableM, nil, scopes
}

// GetVariableParent 获取上一级被覆盖的变量
func GetVariableParent(dbSess *db.Session, name, scope, variableType string, scopes []string, orgId, projectId, tplId models.Id) (bool, models.Variable) {
	variable := models.Variable{}
	query := dbSess.Where("org_id = ?", orgId).
		Where("name = ?", name).
		Where("scope != ?", scope).
		Where("scope in (?)", scopes).
		Where("type = ?", variableType)

	// 只有环境层级需要很细粒度的数据隔离
	if scope == consts.ScopeEnv {
		variables := make([]models.Variable, 0)
		if err := query.Order("scope asc").Find(&variables); err != nil {
			return false, variable
		}
		for _, v := range variables {
			if v.ProjectId != "" {
				if projectId == v.ProjectId {
					return true, v
				}
				continue
			}

			if v.TplId != "" {
				if tplId == v.TplId {
					return true, v
				}
				continue
			}

			if v.TplId == "" && v.ProjectId == "" && v.EnvId == "" {
				return true, v
			}

		}
		return false, variable
	}

	if err := query.
		Order("scope desc").
		First(&variable); err != nil {
		return false, variable
	}

	return true, variable
}
