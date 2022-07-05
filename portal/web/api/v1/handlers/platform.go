// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
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
// @router /platform/stat/basedata [get]
// @Success 200 {object} ctx.JSONResult{result=resps.PlatformBasedataResp}
func (Platform) PlatformStatBasedata(c *ctx.GinRequest) {
	c.JSONResult(apps.PlatformStatBasedata(c.Service()))
}

// PlatformStatProEnv provider环境数量统计
// @Tags 平台
// @Summary provider环境数量统计
// @Description provider环境数量统计
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @router /platform/stat/provider/env [get]
// @Success 200 {object} ctx.JSONResult{result=[]resps.PfProEnvStatResp}
func (Platform) PlatformStatProEnv(c *ctx.GinRequest) {
	c.JSONResult(apps.PlatformStatProEnv(c.Service()))
}
