package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
)

// PlatformStatPg 合规策略组数量
func PlatformStatPg(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	orgIds := parseOrgIds(form.OrgIds)

	count, err := services.GetPolicyGroupCount(c.DB(), orgIds)
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return &resps.PfPgStatResp{
		Name:  "合规策略组数量",
		Count: count,
	}, nil
}

// PlatformStatPolicy 合规策略数量
func PlatformStatPolicy(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	orgIds := parseOrgIds(form.OrgIds)

	count, err := services.GetPolicyCount(c.DB(), orgIds)
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return &resps.PfPgStatResp{
		Name:  "合规策略数量",
		Count: count,
	}, nil
}

// PlatformStatPgStackEnabled 开启合规并绑定策略组的 Stack 数量
func PlatformStatPgStackEnabled(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {
	orgIds := parseOrgIds(form.OrgIds)

	count, err := services.GetPGStackEnabledCount(c.DB(), orgIds)
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}

	return &resps.PfPgStatResp{
		Name:  "开启合规并绑定策略组的Stack数量",
		Count: count,
	}, nil
}

// PlatformStatPgEnvEnabledActivate 开启合规并绑定策略组的活跃环境数量
func PlatformStatPgEnvEnabledActivate(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {

	return nil, nil
}

// PlatformStatPgStackNG 合规不通过的 Stack 数量
func PlatformStatPgStackNG(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {

	return nil, nil
}

// PlatformStatPgEnvNGActivate 合规不通过的活跃环境数量
func PlatformStatPgEnvNGActivate(c *ctx.ServiceContext, form *forms.PfStatForm) (interface{}, e.Error) {

	return nil, nil
}
