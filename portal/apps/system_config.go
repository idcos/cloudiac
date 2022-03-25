// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/configs"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"fmt"
	"strconv"
)

func SearchSystemConfig(c *ctx.ServiceContext) (interface{}, e.Error) {
	rs := make([]resps.SearchSystemConfigResp, 0)
	err := services.QuerySystemConfig(c.DB()).Find(&rs)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	for index, cfg := range rs {
		if cfg.Name == models.SysCfgNameTaskStepTimeout {
			timeoutInSecond, err := strconv.Atoi(cfg.Value)
			if err != nil {
				return nil, e.New(e.InternalError, err)
			}
			rs[index].Value = strconv.Itoa(timeoutInSecond / 60)
		}
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

func GetRegistryAddr(c *ctx.ServiceContext) (interface{}, e.Error) {
	cfg, err := services.GetSystemConfigByName(c.DB(), models.SysCfgNamRegistryAddr)
	var cfgdb = ""
	if err == nil {
		cfgdb = cfg.Value
	}

	return &resps.RegistryAddrResp{
		RegistryAddrFromDB:  cfgdb,
		RegistryAddrFromCfg: configs.Get().RegistryAddr,
	}, nil
}

func UpsertRegistryAddr(c *ctx.ServiceContext, form *forms.RegistryAddrForm) (interface{}, e.Error) {

	cfg, err := services.UpsertRegistryAddr(c.DB(), form.RegistryAddr)
	var cfgdb = ""
	if err == nil {
		cfgdb = cfg.Value
	}

	return &resps.RegistryAddrResp{
		RegistryAddrFromDB:  cfgdb,
		RegistryAddrFromCfg: configs.Get().RegistryAddr,
	}, nil
}

func GetRegistryAddrStr(c *ctx.ServiceContext) string {
	return services.GetRegistryAddrStr(c.DB())
}
