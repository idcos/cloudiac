// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/services"
)

// PlatformStatBasedata 平台基础信息统计
func PlatformStatBasedata(c *ctx.ServiceContext) (interface{}, e.Error) {
	return services.GetBaseDataCount(c.DB())
}

func PlatformStatProEnv(c *ctx.ServiceContext) (interface{}, e.Error) {
	return services.GetProviderEnvCount(c.DB())
}
