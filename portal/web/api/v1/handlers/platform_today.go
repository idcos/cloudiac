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
// @Param form formData forms.PfStatForm true "parameter"
// @router /platform/stat/today/org [get]
// @Success 200 {object} ctx.JSONResult{result=resps.PfTodayStatResp}
func (Platform) PlatformStatTodayOrg(c *ctx.GinRequest) {
	form := forms.PfStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.PlatformStatTodayOrg(c.Service(), &form))
}
