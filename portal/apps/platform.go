// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
)

// PlatformStatBasedata 平台基础信息统计
func PlatformStatBasedata(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	return services.GetBaseDataCount(c.DB())
}

// PlatformStatProEnv provider环境数量统计
func PlatformStatProEnv(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	return services.GetProviderEnvCount(c.DB())
}

// PlatformStatProRes provider资源数量占比
func PlatformStatProRes(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	return services.GetProviderResCount(c.DB())
}
