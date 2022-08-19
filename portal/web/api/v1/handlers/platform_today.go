// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// PlatformStatBasedata 当日新建组织数
// @Tags 平台
// @Summary 当日新建组织数
// @Description 当日新建组织数
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @router /platform/stat/today/org [get]
// @Success 200 {object} ctx.JSONResult{result=resps.PfTodayStatResp}
func (Platform) PlatformStatTodayOrg(c *ctx.GinRequest) {
	c.JSONResult(apps.PlatformStatTodayOrg(c.Service()))
}

// PlatformStatTodayProject 当日新建项目数
// @Tags 平台
// @Summary 当日新建项目数
// @Description 当日新建项目数
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/today/project [get]
// @Success 200 {object} ctx.JSONResult{result=resps.PfTodayStatResp}
func (Platform) PlatformStatTodayProject(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatTodayProject(c.Service(), &form))
}

// PlatformStatTodayStack 当日新建 Stack 数
// @Tags 平台
// @Summary 当日新建 Stack 数
// @Description 当日新建 Stack 数
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/today/template [get]
// @Success 200 {object} ctx.JSONResult{result=resps.PfTodayStatResp}
func (Platform) PlatformStatTodayStack(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatTodayStack(c.Service(), &form))
}

// PlatformStatTodayPG 当日新建合规策略组数量
// @Tags 平台
// @Summary 当日新建合规策略组数量
// @Description 当日新建合规策略组数量
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/today/pg [get]
// @Success 200 {object} ctx.JSONResult{result=resps.PfTodayStatResp}
func (Platform) PlatformStatTodayPG(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatTodayPG(c.Service(), &form))
}

// PlatformStatTodayEnv 当日新建环境数
// @Tags 平台
// @Summary 当日新建环境数
// @Description 当日新建环境数
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/today/env [get]
// @Success 200 {object} ctx.JSONResult{result=resps.PfTodayStatResp}
func (Platform) PlatformStatTodayEnv(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatTodayEnv(c.Service(), &form))
}

// PlatformStatTodayDestroyedEnv 当日销毁环境数
// @Tags 平台
// @Summary 当日销毁环境数
// @Description 当日销毁环境数
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/today/destroyed_env [get]
// @Success 200 {object} ctx.JSONResult{result=resps.PfTodayStatResp}
func (Platform) PlatformStatTodayDestroyedEnv(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatTodayDestroyedEnv(c.Service(), &form))
}

// PlatformStatTodayResType 当日新建资源数：资源类型、数量
// @Tags 平台
// @Summary 当日新建资源数：资源类型、数量
// @Description 当日新建资源数：资源类型、数量
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/today/res_type [get]
// @Success 200 {object} ctx.JSONResult{result=resps.PfTodayResTypeStatResp}
func (Platform) PlatformStatTodayResType(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatTodayResType(c.Service(), &form))
}
