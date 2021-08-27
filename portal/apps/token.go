// Copyright 2021 CloudJ Company Limited. All rights reserved.

package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"cloudiac/utils"
	"cloudiac/utils/logs"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

func SearchToken(c *ctx.ServiceContext, form *forms.SearchTokenForm) (interface{}, e.Error) {
	//todo 鉴权
	query := services.QueryToken(c.DB(), consts.TokenApi)
	query = query.Where("org_id = ?", c.OrgId)
	if form.Status != "" {
		query = query.Where("status = ?", form.Status)
	}
	if form.Q != "" {
		qs := "%" + form.Q + "%"
		query = query.Where("description LIKE ?", qs)
	}

	query = query.Order("created_at DESC")
	rs, err := getPage(query, form, models.Token{})
	if err != nil {
		c.Logger().Errorf("error get page, err %s", err)
		return nil, err
	}
	return rs, nil
}

func CreateToken(c *ctx.ServiceContext, form *forms.CreateTokenForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create token for user %s", c.UserId))
	var (
		expiredAt models.Time
		er        error
	)

	tokenStr, _ := utils.GetUUID()
	if form.ExpiredAt != "" {
		expiredAt, er = models.Time{}.Parse(form.ExpiredAt)
		if er != nil {
			return nil, e.New(e.BadParam, http.StatusBadRequest, er)
		}
	}

	token, err := services.CreateToken(c.DB().Debug(), models.Token{
		Key:         string(tokenStr),
		Type:        form.Type,
		OrgId:       c.OrgId,
		Role:        form.Role,
		ExpiredAt:   expiredAt,
		Description: form.Description,
		CreatorId:   c.UserId,
		EnvId:       form.EnvId,
		Action:      form.Action,
	})
	if err != nil && err.Code() == e.TokenAlreadyExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error creating token, err %s", err)
		return nil, e.AutoNew(err, e.DBError)
	}

	return token, nil
}

func UpdateToken(c *ctx.ServiceContext, form *forms.UpdateTokenForm) (token *models.Token, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update token %s", form.Id))
	if form.Id == "" {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}

	attrs := models.Attrs{}
	if form.HasKey("status") {
		attrs["status"] = form.Status
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}

	token, err = services.UpdateToken(c.DB(), form.Id, attrs)
	if err != nil && err.Code() == e.TokenAliasDuplicate {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error update token, err %s", err)
		return nil, err
	}
	return
}

func DeleteToken(c *ctx.ServiceContext, form *forms.DeleteTokenForm) (result interface{}, re e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete token %s", form.Id))
	if err := services.DeleteToken(c.DB(), form.Id); err != nil {
		return nil, err
	}

	return
}

func DetailTriggerToken(c *ctx.ServiceContext, form *forms.DetailTriggerTokenForm) (result interface{}, re e.Error) {
	token, err := services.DetailTriggerToken(c.DB(), c.OrgId, form.EnvId, form.Action)
	if err != nil {
		// 如果不存在直接返回
		if err.Code() == e.TokenNotExists {
			return token, nil
		}
		return nil, err
	}
	return token, nil
}

func ApiTriggerHandler(c *ctx.ServiceContext, form forms.ApiTriggerHandler) (interface{}, e.Error) {
	var (
		err      e.Error
		taskType string
	)
	tx := c.Tx().Debug()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	token, err := services.IsExistsTriggerToken(tx, form.Token)
	if err != nil {
		_ = tx.Rollback()
		logs.Get().Errorf("get token by envId err %s:", err)
		if err.Code() == e.TokenNotExists {
			return nil, e.New(err.Code(), err, http.StatusForbidden)
		}
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	env, err := services.GetEnvById(tx, token.EnvId)
	if err != nil {
		_ = tx.Rollback()
		logs.Get().Errorf("get env by id err %s:", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	tpl, err := services.GetTemplateById(tx, env.TplId)
	if err != nil {
		_ = tx.Rollback()
		logs.Get().Errorf("get tpl by id err %s:", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	switch token.Action {
	case models.TaskTypePlan:
		taskType = models.TaskTypePlan
	case models.TaskTypeApply:
		taskType = models.TaskTypeApply
	case models.TaskTypeDestroy:
		taskType = models.TaskTypeDestroy
	default:
		return nil, e.New(e.BadRequest, errors.New("token action illegal"), http.StatusBadRequest)
	}

	vars, err, _ := services.GetValidVariables(tx, consts.ScopeEnv, env.OrgId, env.ProjectId, env.TplId, env.Id, true)
	if err != nil {
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	taskVars := services.GetVariableBody(vars)

	task := models.Task{
		Name:        models.Task{}.GetTaskNameByType(taskType),
		Targets:     models.StrSlice{},
		CreatorId:   c.UserId,
		KeyId:       env.KeyId,
		Variables:   taskVars,
		AutoApprove: env.AutoApproval,
		BaseTask: models.BaseTask{
			Type:        taskType,
			Flow:        models.TaskFlow{},
			StepTimeout: env.Timeout,
			RunnerId:    env.RunnerId,
		},
	}

	_, err = services.CreateTask(tx, tpl, env, task)

	if err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error creating task, err %s", err)
		return nil, e.New(err.Code(), err, http.StatusInternalServerError)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		c.Logger().Errorf("error create task, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return nil, nil
}
