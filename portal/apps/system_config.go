// Copyright 2021 CloudJ Company Limited. All rights reserved.

package apps

import (
	"cloudiac/configs"
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

func SearchSystemConfig(c *ctx.ServiceContext) (interface{}, e.Error) {
	rs := make([]SearchSystemConfigResp, 0)
	err := services.QuerySystemConfig(c.DB()).Find(&rs)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return rs, nil
}

func UpdateSystemConfig(c *ctx.ServiceContext, form *forms.UpdateSystemConfigForm) (cfg *models.SystemCfg, err e.Error) {
	tx := c.Tx()
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	for _, v := range form.SystemCfg {
		attrs := models.Attrs{}
		attrs["value"] = v.Value
		if _, err := services.UpdateSystemConfig(tx, v.Name, attrs); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, e.New(e.DBError, fmt.Errorf("error commit database, err %s", err))
	}

	return nil, nil
}

func CheckRegistryAddr(c *ctx.ServiceContext) (interface{}, e.Error) {
	// check db
	cfg, err := services.GetSystemConfigByName(c.DB(), models.SysCfgNamRegistryAddr)
	if err == nil && cfg != nil {
		return map[string]interface{}{
			"isExisted": true,
		}, nil
	}

	// check config file
	if configs.Get().RegistryAddr != "" {
		return map[string]interface{}{
			"isExisted": true,
		}, nil
	}

	return map[string]interface{}{
		"isExisted": false,
	}, nil
}
