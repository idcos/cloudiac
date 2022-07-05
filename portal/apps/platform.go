// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"strings"
)

func parseOrgIds(orgIds string) []string {
	orgIds = strings.Trim(orgIds, " ")
	if strings.Trim(orgIds, " ") == "" {
		return nil
	}

	return strings.Split(orgIds, ",")
}

// PlatformStatBasedata 平台基础信息统计
func PlatformStatBasedata(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	var err error
	var result = &resps.PfBasedataResp{}

	orgIds := parseOrgIds(form.OrgIds)
	dbSess := c.DB()

	// organization
	result.OrgCount.Total, result.OrgCount.Active, err = services.GetOrgTotalAndActiveCount(dbSess, orgIds)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	// project
	result.ProjectCount.Total, result.ProjectCount.Active, err = services.GetProjectTotalAndActiveCount(dbSess, orgIds)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	// enviroment
	result.EnvCount.Total, result.EnvCount.Active, err = services.GetEnvTotalAndActiveCount(dbSess, orgIds)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	// stack
	result.StackCount.Total, result.StackCount.Active, err = services.GetStackTotalAndActiveCount(dbSess, orgIds)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	// user

	return result, nil
}

// PlatformStatProEnv provider环境数量统计
func PlatformStatProEnv(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	return services.GetProviderEnvCount(c.DB())
}

// PlatformStatProRes provider资源数量占比
func PlatformStatProRes(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	return services.GetProviderResCount(c.DB())
}
