// Copyright 2021 CloudJ Company Limited. All rights reserved.

package services

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/db"
	"cloudiac/portal/models"
)

func DeletePolicyGroupRel(tx *db.Session, id models.Id, scope string) e.Error {
	sql := ""
	if scope == consts.ScopeEnv {
		sql = "env_id = ? and group_id != ''"
	} else {
		sql = "tpl_id = ? and env_id = '' and group_id != ''"
	}
	if _, err := tx.Where(sql, id).Delete(models.PolicyRel{}); err != nil {
		if e.IsRecordNotFound(err) {
			return nil
		}
		return e.New(e.DBError, err)
	}
	return nil
}

func GetPolicyRel(query *db.Session, id models.Id, scope string) (*models.PolicyRel, e.Error) {
	sql := ""
	if scope == consts.ScopeEnv {
		sql = "env_id = ? and group_id = ''"
	} else {
		sql = "tpl_id = ? and group_id = ''"
	}
	rel := models.PolicyRel{}
	if err := query.Model(models.PolicyRel{}).Where(sql, id).First(&rel); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.PolicyRelNotExist, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &rel, nil
}

func CreatePolicyRel(tx *db.Session, rel *models.PolicyRel) (*models.PolicyRel, e.Error) {
	if err := models.Create(tx, rel); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.PolicyRelAlreadyExist, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return rel, nil
}

func DeletePolicyEnabledRel(tx *db.Session, id models.Id, scope string) e.Error {
	sql := ""
	if scope == consts.ScopeEnv {
		sql = "env_id = ? and group_id = ''"
	} else {
		sql = "tpl_id = ? and env_id = '' and group_id = ''"
	}
	if _, err := tx.Where(sql, id).Delete(models.PolicyRel{}); err != nil {
		if e.IsRecordNotFound(err) {
			return nil
		}
		return e.New(e.DBError, err)
	}
	return nil
}

// CreateEnvPolicyScan 创建环境关联启动新环境扫描
func CreateEnvPolicyScan(tx *db.Session, tpl *models.Template, env *models.Env) (*models.PolicyRel, e.Error) {
	// 将云模板关联的分组同步到新环境
	groups, err := GetPolicyGroupByTplId(tx, tpl.Id)
	if err != nil {
		// 云模板没有关联策略，环境正常创建
		if err.Code() == e.PolicyGroupNotExist {
			return nil, nil
		}
		return nil, err
	}
	if len(groups) == 0 {
		return nil, nil
	}
	var rels []models.PolicyRel
	for _, group := range groups {
		rels = append(rels, models.PolicyRel{
			OrgId:     env.OrgId,
			ProjectId: env.ProjectId,
			GroupId:   group.Id,
			EnvId:     env.Id,
			Scope:     consts.ScopeEnv,
		})
	}

	if er := models.CreateBatch(tx, rels); er != nil {
		return nil, e.New(e.DBError, er)
	}

	// 同步云模板和环境的合规检测启用状态
	tplRel, err := GetPolicyRel(tx, tpl.Id, consts.ScopeTemplate)
	if err != nil {
		// 云模板合规检测未启用，环境也不启用合规检测
		if err.Code() == e.PolicyRelNotExist {
			return nil, nil
		}
		return nil, e.New(err.Code(), err)
	}

	// 云模板合规检测未启用，环境也不启用合规检测
	if tplRel == nil {
		return nil, nil
	}

	// 启用环境合规检测
	envRel := &models.PolicyRel{
		OrgId:     env.OrgId,
		ProjectId: env.ProjectId,
		EnvId:     env.Id,
		Enabled:   true,
	}

	if _, err := CreatePolicyRel(tx, envRel); err != nil {
		return nil, e.New(err.Code(), err)
	}

	return envRel, nil
}
