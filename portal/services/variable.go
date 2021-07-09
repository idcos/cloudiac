package services

import (
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

func SearchVariable(dbSess *db.Session, orgId, projectId, tplId, envId models.Id, scope string) ([]models.Variable, e.Error) {
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

func OperationVariables(tx *db.Session, orgId, projectId, tplId, envId models.Id, variables []forms.Variables) e.Error {
	bq := utils.NewBatchSQL(1024, "INSERT INTO", models.Variable{}.TableName(),
		"id", "scope", "type", "name", "value", "sensitive", "description", "org_id", "project_id", "tpl_id", "env_id")
	for _, v := range variables {
		var value string
		if v.Sensitive && v.Value != "" {
			value, _ = utils.AesEncrypt(v.Value)
		}
		if v.Value == "" && v.Sensitive {
			value = v.Value
		}
		//id不为空修改变量，反之新建
		if v.Id != "" {
			attrs := map[string]interface{}{
				"name":        v.Name,
				"value":       value,
				"sensitive":   v.Sensitive,
				"description": v.Description,
			}
			err := UpdateVariable(tx, v.Id, attrs)
			if err != nil && err.Code() == e.VariableAliasDuplicate {
				return e.New(err.Code(), err, http.StatusBadRequest)
			} else if err != nil {
				_ = tx.Rollback()
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
	if _, err := models.UpdateModel(tx,
		models.Variable{BaseModel: models.BaseModel{Id: variableId}}, attr); err != nil {
		if e.IsDuplicate(err) {
			return e.New(e.VariableAliasDuplicate)
		}
		return e.New(e.DBError, fmt.Errorf("update variable error: %v", err))
	}
	return nil
}

func DeleteVariables(tx *db.Session, DeleteVariables []string) e.Error {
	if _, err := tx.Where("id in (?)", DeleteVariables).Delete(&models.Variable{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete variables error: %v", err))
	}
	return nil

}
