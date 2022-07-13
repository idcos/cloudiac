// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// PlatformStatPg 合规策略组数量
// @Tags 平台
// @Summary 合规策略组数量
// @Description 合规策略组数量
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/pg [get]
// @Success 200 {object} ctx.JSONResult{result=resps.pfPgStatResp}
func (Platform) PlatformStatPg(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatPg(c.Service(), &form))
}

// PlatformStatPolicy 合规策略数量
// @Tags 平台
// @Summary 合规策略数量
// @Description 合规策略数量
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/pg [get]
// @Success 200 {object} ctx.JSONResult{result=resps.pfPgStatResp}
func (Platform) PlatformStatPolicy(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatPolicy(c.Service(), &form))
}