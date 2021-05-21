package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/libs/db"
	"cloudiac/libs/page"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/utils"
)

func SearchWebhook(c *ctx.ServiceCtx, form *forms.SearchWebhookForm) (interface{}, e.Error) {
	var query *db.Session
	tx := c.DB()
	if form.TplId != 0 {
		query = services.SearchWebhookByTplId(tx, form.TplId)
	}

	if form.TplGuid != "" {
		query = services.SearchWebhookByTplGuid(tx, form.TplGuid)
	}
	if query != nil {
		query = query.Order("created_at DESC")
	}
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	webhookResp := make([]*models.TemplateWebhook, 0)
	if err := p.Scan(&webhookResp); err != nil {
		return nil, e.New(e.DBError, err)
	}
	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     webhookResp,
	}, nil
}

func CreateWebhook(c *ctx.ServiceCtx, form *forms.CreateWebhookForm) (interface{}, e.Error) {
	webhook, err := services.CreateWebhook(c.DB(), models.TemplateWebhook{
		TplGuid:     form.TplGuid,
		TplId:       form.TplId,
		Action:      form.Action,
		AccessToken: utils.GenGuid(""),
	})
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	return webhook, nil
}

func UpdateWebhook(c *ctx.ServiceCtx, form *forms.UpdateWebhookForm) (interface{}, e.Error) {
	attrs := models.Attrs{}
	if form.HasKey("action") {
		attrs["action"] = form.Action
	}
	return services.UpdateWebhook(c.DB(), form.Id, attrs)
}

func DeleteWebhook(c *ctx.ServiceCtx, form *forms.DeleteWebhookForm) (interface{},  e.Error) {
	return services.DeleteWebhook(c.DB(), form.Id)
}

func DetailWebhook(c *ctx.ServiceCtx, form *forms.DetailWebhookForm) (interface{},  e.Error) {
	return services.DetailWebhook(c.DB(), form.Id)
}