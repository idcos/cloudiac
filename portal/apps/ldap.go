// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
)

func GetLdapOUs(c *ctx.ServiceContext) (interface{}, e.Error) {
	ous, err := services.SearchLdapOUs()
	if err != nil {
		return nil, err
	}

	var resp = &resps.LdapOUListResp{
		LdapOUs: ous,
	}

	return resp, nil
}

func GetLdapUsers(c *ctx.ServiceContext, form *forms.SearchLdapUserForm) (interface{}, e.Error) {
	return nil, nil
}

func AuthLdapUser(c *ctx.ServiceContext, form *forms.AuthLdapUserForm) (interface{}, e.Error) {
	return nil, nil
}

func AuthLdapOU(c *ctx.ServiceContext, form *forms.AuthLdapOUForm) (interface{}, e.Error) {
	return nil, nil
}
