package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"fmt"
)

type searchSystemConfigResp struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

func (m *searchSystemConfigResp) TableName() string {
	return models.SystemCfg{}.TableName()
}

func SearchSystemConfig(c *ctx.ServiceCtx) (interface{}, e.Error) {
	rs := []*searchSystemConfigResp{}
	err := services.QuerySystemConfig(c.DB()).Find(&rs)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return rs, nil
}

func UpdateSystemConfig(c *ctx.ServiceCtx, form *forms.UpdateSystemConfigForm) (cfg *models.SystemCfg, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update system config %d", form.Id))
	if form.Id == 0 {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}

	attrs := models.Attrs{}
	if form.HasKey("value") {
		attrs["value"] = form.Value
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}

	cfg, err = services.UpdateSystemConfig(c.DB(), form.Id, attrs)
	return
}
