// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"cloudiac/portal/services/vcsrv"
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
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	tokens := make([]models.Token, 0)
	if err := p.Scan(&tokens); err != nil {
		return nil, e.New(e.DBError, err)
	}

	var tokenResp = make([]resps.TokenResp, 0)
	for _, token := range tokens {
		tokenResp = append(tokenResp, resps.TokenResp{
			TimedModel:  token.TimedModel,
			Name:        token.Name,
			Type:        token.Type,
			OrgId:       token.OrgId,
			Role:        token.Role,
			Status:      token.Status,
			ExpiredAt:   token.ExpiredAt,
			Description: token.Description,
			CreatorId:   token.CreatorId,
			EnvId:       token.EnvId,
			Action:      token.Action,
		})
	}

	return &page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     tokenResp,
	}, nil
}

func CreateToken(c *ctx.ServiceContext, form *forms.CreateTokenForm) (*models.Token, e.Error) {
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

	token, err := services.CreateToken(c.DB(), models.Token{
		Name:        form.Name,
		Key:         string(tokenStr),
		Type:        form.Type,
		OrgId:       c.OrgId,
		Role:        form.Role,
		ExpiredAt:   &expiredAt,
		Description: form.Description,
		CreatorId:   c.UserId,
		//EnvId:       form.EnvId,
		//Action:      form.Action,
	})
	if err != nil && err.Code() == e.TokenAlreadyExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error creating token, err %s", err)
		return nil, e.AutoNew(err, e.DBError)
	}

	return token, nil
}

func UpdateToken(c *ctx.ServiceContext, form *forms.UpdateTokenForm) (tokenResp *resps.TokenResp, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update token %s", form.Id))
	if form.Id == "" {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("status") {
		attrs["status"] = form.Status
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}

	token, err := services.UpdateToken(c.DB(), form.Id, attrs)
	if err != nil && err.Code() == e.TokenAliasDuplicate {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error update token, err %s", err)
		return nil, err
	}

	return &resps.TokenResp{
		TimedModel:  token.TimedModel,
		Name:        token.Name,
		Type:        token.Type,
		OrgId:       token.OrgId,
		Role:        token.Role,
		Status:      token.Status,
		ExpiredAt:   token.ExpiredAt,
		Description: token.Description,
		CreatorId:   token.CreatorId,
		EnvId:       token.EnvId,
		Action:      token.Action,
	}, nil
}

func DeleteToken(c *ctx.ServiceContext, form *forms.DeleteTokenForm) (result interface{}, re e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete token %s", form.Id))
	if err := services.DeleteToken(c.DB(), form.Id); err != nil {
		return nil, err
	}

	return
}

func VcsWebhookUrl(c *ctx.ServiceContext, form *forms.VcsWebhookUrlForm) (result interface{}, re e.Error) {
	token, err := GetWebhookToken(c)
	if err != nil {
		return nil, err
	}

	tpl, err := services.GetTplByEnvId(c.DB(), form.EnvId)
	if err != nil {
		return nil, err
	}

	vcs, err := services.GetVcsById(c.DB(), tpl.VcsId)
	if err != nil {
		return nil, err
	}

	webhookUrl := vcsrv.GetWebhookUrl(vcs, token.Key)
	return struct {
		Url string `json:"url"`
	}{webhookUrl}, err
}

func ApiTriggerHandler(c *ctx.ServiceContext, form forms.ApiTriggerHandler) (interface{}, e.Error) {
	var (
		err      e.Error
		taskType string
	)
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	token, err := services.IsActiveToken(tx, form.Token, consts.TokenTrigger)
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
	//todo 合规
	default:
		return nil, e.New(e.BadRequest, errors.New("token action illegal"), http.StatusBadRequest)
	}

	// 计算变量列表
	vars, er := services.GetValidVarsAndVgVars(tx, env.OrgId, env.ProjectId, env.TplId, env.Id)
	if er != nil {
		_ = tx.Rollback()
		return nil, err
	}

	task := models.Task{
		Name:        models.Task{}.GetTaskNameByType(taskType),
		Targets:     models.StrSlice{},
		CreatorId:   consts.SysUserId,
		KeyId:       env.KeyId,
		Variables:   vars,
		AutoApprove: env.AutoApproval,
		BaseTask: models.BaseTask{
			Type:        taskType,
			StepTimeout: env.StepTimeout,
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

func GetWebhookToken(c *ctx.ServiceContext) (*models.Token, e.Error) {
	// 获取token
	var (
		token *models.Token
		err   e.Error
	)

	token, err = services.DetailTriggerToken(c.DB(), c.OrgId)
	if err != nil && err.Code() == e.TokenNotExists {
		// 如果不存在, 则创建一个触发器token
		token, err = CreateToken(c, &forms.CreateTokenForm{
			Type: consts.TokenTrigger,
		})
	}
	return token, err
}
