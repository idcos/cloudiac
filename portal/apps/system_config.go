package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
	"fmt"
)

type SearchSystemConfigResp struct {
	Id          models.Id `json:"id"`
	Name        string    `json:"name"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
}

func (m *SearchSystemConfigResp) TableName() string {
	return models.SystemCfg{}.TableName()
}

func SearchSystemConfig(c *ctx.ServiceCtx) (interface{}, e.Error) {
	rs := SearchSystemConfigResp{}
	err := services.QuerySystemConfig(c.DB()).First(&rs)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return rs, nil
}

func UpdateSystemConfig(c *ctx.ServiceCtx, form *forms.UpdateSystemConfigForm) (cfg *models.SystemCfg, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update system config %s", form.Id))

	attrs := models.Attrs{}
	if form.HasKey("value") {
		attrs["value"] = form.Value
	}

	cfg, err = services.UpdateSystemConfig(c.DB(), attrs)
	return
}
