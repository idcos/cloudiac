package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
	"net/http"
)

var (
	emailSubjectResetPassword = "密码重置通知【CloudIaC】"
	emailBodyResetPassword    = "尊敬的 {{.Name}}：\n\n您的密码已经被重置，这是您的新密码：\n\n密码：\t{{.InitPass}}\n\n请使用新密码登陆系统。\n\n为了保障您的安全，请立即登陆您的账号并修改密码。"
)

func SearchToken(c *ctx.ServiceCtx, form *forms.SearchTokenForm) (interface{}, e.Error) {
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

func CreateToken(c *ctx.ServiceCtx, form *forms.CreateTokenForm) (interface{}, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create token for user %s", c.UserId))

	tokenStr := models.NewId("")
	token, err := services.CreateToken(c.DB().Debug(), models.Token{
		Key:         string(tokenStr),
		Type:        form.Type,
		OrgId:       c.OrgId,
		Role:        form.Role,
		ExpiredAt:   form.ExpiredAt,
		Description: form.Description,
		CreatorId:   c.UserId,
	})
	if err != nil && err.Code() == e.TokenAlreadyExists {
		return nil, e.New(err.Code(), err, http.StatusBadRequest)
	} else if err != nil {
		c.Logger().Errorf("error creating token, err %s", err)
		return nil, e.AutoNew(err, e.DBError)
	}

	return token, nil
}

func UpdateToken(c *ctx.ServiceCtx, form *forms.UpdateTokenForm) (token *models.Token, err e.Error) {
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

func DeleteToken(c *ctx.ServiceCtx, form *forms.DeleteTokenForm) (result interface{}, re e.Error) {
	c.AddLogField("action", fmt.Sprintf("delete token %s", form.Id))
	if err := services.DeleteToken(c.DB(), form.Id); err != nil {
		return nil, err
	}

	return
}
