// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Platform struct {
	ctrl.GinController
}

// PlatformStatBasedata 平台基础信息统计接口
// @Tags 平台
// @Summary 平台基础信息统计接口
// @Description 平台基础信息统计接口
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/basedata [get]
// @Success 200 {object} ctx.JSONResult{result=resps.PfBasedataResp}
func (Platform) PlatformStatBasedata(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatBasedata(c.Service(), &form))
}

// PlatformStatProEnv provider环境数量统计
// @Tags 平台
// @Summary provider环境数量统计
// @Description provider环境数量统计
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/provider/env [get]
// @Success 200 {object} ctx.JSONResult{result=[]resps.PfProEnvStatResp}
func (Platform) PlatformStatProEnv(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatProEnv(c.Service(), &form))
}

// PlatformStatProRes provider资源数量占比
// @Tags 平台
// @Summary provider资源数量占比
// @Description provider资源数量占比
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/provider/resource [get]
// @Success 200 {object} ctx.JSONResult{result=[]resps.PfProResStatResp}
func (Platform) PlatformStatProRes(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatProRes(c.Service(), &form))
}

// PlatformStatResType 资源类型占比
// @Tags 平台
// @Summary 资源类型占比
// @Description 资源类型占比
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/resource/type [get]
// @Success 200 {object} ctx.JSONResult{result=[]resps.PfResTypeStatResp}
func (Platform) PlatformStatResType(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatResType(c.Service(), &form))
}

// PlatformStatResWeekChange 一周资源变更趋势
// @Tags 平台
// @Summary 一周资源变更趋势
// @Description 一周资源变更趋势
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/resource/week [get]
// @Success 200 {object} ctx.JSONResult{result=[]resps.PfResWeekChangeResp}
func (Platform) PlatformStatResWeekChange(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatResWeekChange(c.Service(), &form))
}

// PlatformStatActiveResType 活跃资源数量
// @Tags 平台
// @Summary 活跃资源数量
// @Description 活跃资源数量
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/resource/active [get]
// @Success 200 {object} ctx.JSONResult{result=resps.PfActiveResStatResp}
func (Platform) PlatformStatActiveResType(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatActiveResType(c.Service(), &form))
}

// PlatformOperationLog 操作日志
// @Tags 平台
// @Summary 操作日志
// @Description 操作日志
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/operation/log [get]
// @Success 200 {object} ctx.JSONResult{result=resps.OperationLogResp}
func (Platform) PlatformOperationLog(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformOperationLog(c.Service(), &form))
}
