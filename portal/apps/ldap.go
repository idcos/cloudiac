// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
)

func GetLdapOUs(c *ctx.ServiceContext) (interface{}, e.Error) {
	ous, err := services.SearchLdapOUs()
	return ous, err
}

// TODO: 未过滤用户，前端过滤，返回所有用户
func GetLdapUsers(c *ctx.ServiceContext, form *forms.SearchLdapUserForm) (interface{}, e.Error) {
	users, err := services.SearchLdapUsers(form.Q, 0)
	if err != nil {
		return nil, err
	}

	var resp = &resps.LdapUserListResp{
		LdapUsers: users,
	}

	return resp, nil
}

func AuthLdapUser(c *ctx.ServiceContext, form *forms.AuthLdapUserForm) (interface{}, e.Error) {
	result, err := services.CreateLdapUserOrg(c.DB(), c.OrgId, models.User{
		Name:  form.Uid,
		Email: form.Email,
		Phone: form.Phone,
	}, form.Role)

	return result, err
}

func AuthLdapOU(c *ctx.ServiceContext, form *forms.AuthLdapOUForm) (interface{}, e.Error) {
	result, err := services.CreateOUOrg(c.DB(), models.LdapOUOrg{
		OrgId: c.OrgId,
		DN:    form.DN,
		OU:    form.OU,
		Role:  form.Role,
	})

	return result, err
}
