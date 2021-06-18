package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/libs/page"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/utils"
)

func SearchAccessToken(c *ctx.ServiceCtx, form *forms.SearchAccessTokenForm) (interface{}, e.Error) {
	query := services.SearchAccessTokenByTplGuid(c.DB(), form.TplGuid).Order("created_at DESC")
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	webhookResp := make([]*models.TemplateAccessToken, 0)
	if err := p.Scan(&webhookResp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     webhookResp,
	}, nil
}

func CreateAccessToken(c *ctx.ServiceCtx, form *forms.CreateAccessTokenForm) (interface{}, e.Error) {
	webhook, err := services.CreateAccessToken(c.DB(), models.TemplateAccessToken{
		TplGuid:     form.TplGuid,
		Action:      form.Action,
		AccessToken: utils.GenGuid(""),
	})
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	return webhook, nil
}

func UpdateAccessToken(c *ctx.ServiceCtx, form *forms.UpdateAccessTokenForm) (interface{}, e.Error) {
	attrs := models.Attrs{}
	if form.HasKey("action") {
		attrs["action"] = form.Action
	}
	return services.UpdateAccessToken(c.DB(), form.Id, attrs)
}

func DeleteAccessToken(c *ctx.ServiceCtx, form *forms.DeleteAccessTokenForm) (interface{}, e.Error) {
	return services.DeleteAccessToken(c.DB(), form.Id)
}

func DetailAccessToken(c *ctx.ServiceCtx, form *forms.DetailAccessTokenForm) (interface{}, e.Error) {
	return services.DetailAccessToken(c.DB(), form.Id)
}
