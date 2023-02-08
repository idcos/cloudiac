// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts"
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
	"fmt"
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
	result.UserCount.Total, result.UserCount.Active, err = services.GetUserTotalAndActiveCount(dbSess, orgIds)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return result, nil
}

// PlatformStatProEnv provider环境数量统计
func PlatformStatProEnv(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	orgIds := parseOrgIds(form.OrgIds)

	return services.GetProviderEnvCount(c.DB(), orgIds)
}

// PlatformStatProRes provider资源数量占比
func PlatformStatProRes(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	orgIds := parseOrgIds(form.OrgIds)
	return services.GetProviderResCount(c.DB(), orgIds)
}

// PlatformStatResType 资源类型占比
func PlatformStatResType(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	orgIds := parseOrgIds(form.OrgIds)
	return services.GetResTypeCount(c.DB(), orgIds)
}

// PlatformStatActiveResType 活跃资源数量
func PlatformStatActiveResType(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	orgIds := parseOrgIds(form.OrgIds)
	return services.GetOrgActiveResTypeCount(c.DB(), orgIds)
}

// PlatformStatResWeekChange 一周资源变更趋势
func PlatformStatResWeekChange(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	orgIds := parseOrgIds(form.OrgIds)
	return services.GetResWeekChange(c.DB(), orgIds)
}

// PlatformOperationLog 操作日志
func PlatformOperationLog(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	query := services.PlatformOperationLog(c.DB(), parseOrgIds(form.OrgIds))
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	result := make([]resps.OperationLogResp, 0)
	if err := p.Scan(&result); err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	for index, v := range result {
		result[index].ActionName = consts.UserOperationLogAttr[fmt.Sprintf("%s.%s", v.ObjectType, v.Action)]
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     result,
	}, nil
}
