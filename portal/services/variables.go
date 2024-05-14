// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package services

import (
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
		variable.Id = variable.NewId()
	}
	if err := models.Create(tx, &variable); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.VariableAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &variable, nil
}

func SearchVariable(dbSess *db.Session, orgId, projectId, envId models.Id) ([]models.Variable, e.Error) {
	variables := make([]models.Variable, 0)
	if err := dbSess.Model(&models.Variable{}).
		//Where("org_id = ?", orgId).
		Where("org_id = ? AND project_id = ? AND env_id = ?", orgId, " ", " ").Or("project_id = ? AND env_id = ?", projectId, " ").Or("env_id = ?", envId).
		// 按照枚举值排序控制org类型在最上面
		Order("scope asc").
		Find(&variables); err != nil {
		return nil, e.New(e.DBError, err)
	} //nolint
	return variables, nil
}

func SearchVariableByTemplateId(tx *db.Session, tplId models.Id) ([]models.Variable, e.Error) {
	variables := make([]models.Variable, 0)
	if err := tx.Where("tpl_id = ? AND scope = ?", tplId, "template").Find(&variables); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return variables, nil
}

func OperationVariables(tx *db.Session, orgId, projectId, tplId, envId models.Id,
	variables []forms.Variable, deleteVariablesId []string) e.Error {
	if err := deleteVariables(tx, deleteVariablesId); err != nil {
		return err
	}

	bq := utils.NewBatchSQL(1024, "INSERT INTO", models.Variable{}.TableName(),
		"id", "scope", "type", "name", "value", "sensitive", "description", "org_id", "project_id", "tpl_id", "env_id", "options")
	for _, v := range variables {
		attrs := map[string]interface{}{
			"name":        v.Name,
			"sensitive":   v.Sensitive,
			"description": v.Description,
			"options":     v.Options,
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
		if err := updateOrCreateVars(v, attrs, tx, value, bq, []models.Id{orgId, projectId, tplId, envId}); err != nil {
			return err
		}
	}
	if err := createVariables(tx, bq); err != nil {
		return err
	}
	return nil
}

func updateOrCreateVars(v forms.Variable, attrs map[string]interface{}, tx *db.Session, value string, bq *utils.BatchSQL, ids []models.Id) e.Error {
	if v.Id != "" {
		err := updateVariable(tx, v.Id, attrs)
		if err != nil && err.Code() == e.VariableAliasDuplicate {
			return e.New(err.Code(), err, http.StatusBadRequest)
		} else if err != nil {
			return err
		}
	} else {
		vId := models.Variable{}.NewId()
		if err := bq.AddRow(vId, v.Scope, v.Type, v.Name, value, v.Sensitive, v.Description,
			ids[0], ids[1], ids[2], ids[3], v.Options); err != nil {
			return e.New(e.DBError, err)
		}
	}
	return nil
}

func createVariables(tx *db.Session, bq *utils.BatchSQL) e.Error {
	for bq.HasNext() {
		sql, args := bq.Next()
		if _, err := tx.Exec(sql, args...); err != nil {
			return e.New(e.DBError, err)
		}
	}
	return nil
}

func updateVariable(tx *db.Session, variableId models.Id, attr map[string]interface{}) e.Error {
	if _, err := models.UpdateAttr(tx,
		models.Variable{BaseModel: models.BaseModel{Id: variableId}}, attr); err != nil {
		if e.IsDuplicate(err) {
			return e.New(e.VariableAliasDuplicate)
		}
		return e.New(e.DBError, fmt.Errorf("update variable error: %v", err))
	}
	return nil
}

func deleteVariables(tx *db.Session, varIds []string) e.Error {
	if len(varIds) == 0 {
		return nil
	}
	if _, err := tx.Where("id IN (?)", varIds).Delete(&models.Variable{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete variables error: %v", err))
	}
	return nil
}

func GetValidVariables(dbSess *db.Session, scope string, orgId, projectId, tplId, envId models.Id, keepSensitive bool) (map[string]models.Variable, e.Error, []string) {

	// 根据scope 构建变量应用范围
	var scopes []string
	switch scope {
	case consts.ScopeEnv:
		scopes = consts.EnvScopeEnv
	case consts.ScopeTemplate:
		scopes = consts.EnvScopeTpl
	case consts.ScopeProject:
		scopes = consts.EnvScopeProject
	case consts.ScopeOrg:
		scopes = consts.EnvScopeOrg
	default:
		panic(fmt.Errorf("unknown scope '%s'", scope))
	}

	// 将组织下所有的变量查询，在代码处理变量的继承关系及是否要应用该变量
	variables, err := SearchVariable(dbSess, orgId, projectId, envId)
	if err != nil {
		return nil, err, scopes
	}
	variableM := getNewVarsMap(variables, scopes, keepSensitive, projectId, tplId, envId)

	return variableM, nil, scopes
}

func getNewVarsMap(variables []models.Variable, scopes []string, keepSensitive bool, projectId, tplId, envId models.Id) map[string]models.Variable {
	variableM := make(map[string]models.Variable)
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
	return variableM
}

type GetVarParentParams struct {
	OrgId        models.Id
	ProjectId    models.Id
	TplId        models.Id
	Name         string
	Scope        string
	VariableType string
}

func checkVariable(v models.Variable, projectId, tplId models.Id) bool {
	if v.ProjectId != "" {
		if projectId == v.ProjectId {
			return true
		}
	}

	if v.TplId != "" {
		if tplId == v.TplId {
			return true
		}
	}

	if v.TplId == "" && v.ProjectId == "" && v.EnvId == "" {
		return true
	}

	return false
}

// GetVariableParent 获取上一级被覆盖的变量
func GetVariableParent(dbSess *db.Session, scopes []string, varParent GetVarParentParams) (bool, models.Variable) {
	variable := models.Variable{}
	query := dbSess.Where("org_id = ?", varParent.OrgId).
		Where("name = ?", varParent.Name).
		Where("scope != ?", varParent.Scope).
		Where("scope in (?)", scopes).
		Where("type = ?", varParent.VariableType)

	if varParent.Scope != consts.ScopeEnv {
		if err := query.
			Order("scope desc").
			First(&variable); err != nil {
			return false, variable
		} else {
			return true, variable
		}
	}

	// 只有环境层级需要很细粒度的数据隔离
	variables := make([]models.Variable, 0)
	if err := query.Order("scope asc").Find(&variables); err != nil {
		return false, variable
	}
	for _, v := range variables {
		if checkVariable(v, varParent.ProjectId, varParent.TplId) {
			return true, v
		}
	}
	return false, variable

}
func GetVariableBody(vars map[string]models.Variable) []models.VariableBody {
	vb := make([]models.VariableBody, 0, len(vars))
	for k := range vars {
		vb = append(vb, vars[k].VariableBody)
	}
	return vb
}

func QueryVariable(dbSess *db.Session) *db.Session {
	return dbSess.Model(&models.Variable{})
}

func processVarsforUpdate(varsMap map[string]models.Variable, vars []models.Variable) e.Error {
	for i, v := range vars {
		var err error
		if v.Sensitive {
			if v.Value != "" {
				if v.Value, err = utils.EncryptSecretVar(v.Value); err != nil {
					return e.AutoNew(err, e.EncryptError)
				}
				vars[i] = v
			}
		}
		varsMap[v.Key()] = vars[i]
	}
	return nil
}

// UpdateObjectVars 更新(或新增)实例的变量
// vars 参数为待更新的变量,
// 该函数为全量更新，vars 中不存在的变量会从实例的变量列表中删除，己存在的同名变量会被更新，新增变量会创建
func UpdateObjectVars(tx *db.Session, scope string, objectId models.Id, vars []models.Variable) ([]models.Variable, e.Error) {
	table := models.Variable{}.TableName()

	varsMap := make(map[string]models.Variable)
	if err := processVarsforUpdate(varsMap, vars); err != nil {
		return nil, err
	}

	dbVars := make([]models.Variable, 0)
	if err := WithVarScopeIdWhere(tx, table, scope, objectId).Find(&dbVars); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	dbVarsMap := make(map[string]models.Variable)
	for _, v := range dbVars {
		dbVarsMap[v.Key()] = v
	}

	// 删除实例己有，但此次未提交的变量
	delVars := make(map[string][]string)
	for _, v := range dbVars {
		if _, ok := varsMap[v.Key()]; !ok {
			delVars[v.Type] = append(delVars[v.Type], v.Name)
		}
	}
	for typ, names := range delVars {
		if _, err := WithVarScopeIdWhere(tx.Debug(), table, scope, objectId).
			Where("`type` = ? AND name in (?)", typ, names).
			Delete(&models.Variable{}); err != nil {
			return nil, e.AutoNew(err, e.DBError)
		}
	}

	if err := insertVars(dbVarsMap, vars, tx); err != nil {
		return nil, err
	}
	retVars := make([]models.Variable, 0)
	if err := WithVarScopeIdWhere(tx, table, scope, objectId).Find(&retVars); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return retVars, nil
}

func insertVars(dbVarsMap map[string]models.Variable, vars []models.Variable, tx *db.Session) e.Error {
	insertSqls := utils.NewBatchSQL(1024, "INSERT INTO", models.Variable{}.TableName(),
		"org_id", "project_id", "tpl_id", "env_id",
		"id", "scope", "type", "name", "value", "sensitive", "description", "options")

	for _, v := range vars {
		if dbVar, ok := dbVarsMap[v.Key()]; ok { // 同名变量己存在，进行更新
			//敏感变量提交空值表示不进行修改
			if v.Sensitive && v.Value == "" {
				continue
			}

			// 如果没有传 id 则使用己有的 id
			if v.Id == "" {
				v.Id = dbVar.Id
			}
			if _, err := models.UpdateModelAll(tx, v, "id = ?", dbVar.Id); err != nil {
				return e.AutoNew(err, e.DBError)
			}
		} else { // 否则插入新变量
			insertSqls.MustAddRow(v.OrgId, v.ProjectId, v.TplId, v.EnvId,
				v.NewId(), v.Scope, v.Type, v.Name, v.Value, v.Sensitive, v.Description, v.Options)
		}
	}

	for insertSqls.HasNext() {
		sql, args := insertSqls.Next()
		if _, err := tx.Exec(sql, args...); err != nil {
			return e.AutoNew(err, e.DBError)
		}
	}
	return nil
}

func WithVarScopeIdWhere(query *db.Session, tableName string, scope string, id models.Id) *db.Session {
	query = query.Where(fmt.Sprintf("`%s`.`scope` = ?", tableName), scope)
	switch scope {
	case consts.ScopeOrg:
		return query.Where("org_id = ?", id)
	case consts.ScopeProject:
		return query.Where("project_id = ?", id)
	case consts.ScopeTemplate:
		return query.Where("tpl_id = ?", id)
	case consts.ScopeEnv:
		return query.Where("env_id = ?", id)
	default:
		panic(fmt.Errorf("unknown variable scope '%v'", scope))
	}
}

// VarsDesensitization 变量脱敏(将 sensitive 变量的 value 设置为空字符串)
func VarsDesensitization(vars []models.Variable) []models.Variable {
	retVars := make([]models.Variable, 0, len(vars))
	for _, v := range vars {
		if v.Sensitive {
			v.Value = ""
		}
		retVars = append(retVars, v)
	}
	return retVars
}
