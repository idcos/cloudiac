package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/utils"
	"fmt"
)

func CreateVariable(tx *db.Session, variable models.Variable) (*models.Variable, e.Error) {
	if variable.Id == "" {
		variable.Id = models.NewId("v")
	}
	if err := models.Create(tx, &variable); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.DBError, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &variable, nil
}

func SearchVariable(query *db.Session, orgId, projectId, tplId, envId models.Id, scope string) *db.Session {
	query = query.Model(models.Variable{}.TableName())
	switch scope {
	case consts.ScopeOrg:
		query = query.Where("org_id = ?", orgId)
	case consts.ScopeProject:
		query = query.Where("org_id = ?", orgId).
			Where("project_id = ?", projectId)
	case consts.ScopeTemplate:
		query = query.Where("org_id = ?", orgId).
			Where("project_id = ?", projectId).
			Where("template_id = ?", tplId)
	case consts.ScopeEnv:
		query = query.Where("org_id = ?", orgId).
			Where("project_id = ?", projectId).
			Where("template_id = ?", tplId).
			Where("env_id = ?", envId)

	}

	return query.Group("name")
}

func BindVariable(tx *db.Session, orgId, projectId, tplId, envId models.Id, variables []forms.Variables) ([]forms.Variables, e.Error) {
	updateVariables := make([]forms.Variables, 0)
	bq := utils.NewBatchSQL(1024, "INSERT INTO", models.Variable{}.TableName(),
		"scope", "type", "name", "value", "sensitive", "description", "org_id", "project_id", "tpl_id", "env_id")
	for _, v := range variables {
		var value string
		if v.Id != "" {
			updateVariables = append(updateVariables, v)
			continue
		}
		if v.Sensitive {
			value, _ = utils.AesEncrypt(v.Value)
		}
		if err := bq.AddRow(v.Scope, v.Type, v.Name, value, v.Sensitive, v.Description,
			orgId, projectId, tplId, envId); err != nil {
			return updateVariables, e.New(e.DBError, err)
		}
	}

	for bq.HasNext() {
		sql, args := bq.Next()
		if _, err := tx.Exec(sql, args); err != nil {
			return updateVariables, e.New(e.DBError, err)
		}
	}
	return updateVariables, nil
}

func UpdateVariable(tx *db.Session, variableId models.Id, m map[string]interface{}) e.Error {
	if _, err := models.UpdateModel(tx,
		models.Variable{BaseModel: models.BaseModel{Id: variableId}}); err != nil {
		if e.IsDuplicate(err) {
			return e.New(e.VariableAliasDuplicate)
		}
		return e.New(e.DBError, fmt.Errorf("update org error: %v", err))
	}
	return nil
}
