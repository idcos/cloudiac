// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/models/resps"
	"cloudiac/portal/services"
)

func GetLdapOUs(c *ctx.ServiceContext) (interface{}, e.Error) {
	ous, err := services.SearchLdapOUs()
	return ous, err
}

func GetLdapOUsFromDB(c *ctx.ServiceContext, form *forms.SearchLdapOUForm) (interface{}, e.Error) {
	query := c.DB().Model(&models.LdapOUOrg{}).Select("id", "dn", "ou", "role", "created_at")
	p := page.New(form.CurrentPage(), form.PageSize(), query)

	var list = make([]resps.LdapOUDBResp, 0)
	if err := p.Scan(&list); err != nil {
		c.Logger().Errorf("error get ldap ous, err %s", err)
		return nil, e.New(e.DBError, err)
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     list,
	}, nil
}

func DeleteLdapOUFromDB(c *ctx.ServiceContext, form *forms.DeleteLdapOUForm) (interface{}, e.Error) {
	_, err := c.DB().Where(`id = ?`, form.Id).Delete(&models.LdapOUOrg{})
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return resps.DeleteLdapOUResp{
		Id: form.Id,
	}, nil
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

func GetOrgLdapOUs(c *ctx.ServiceContext) (interface{}, e.Error) {
	results, err := services.GetOrgLdapOUs(c.DB(), c.OrgId)
	return &resps.OrgLdapOUListResp{
		OrgLdapOUs: results,
	}, err
}

func DeleteProjectLdapOU(c *ctx.ServiceContext, form *forms.DeleteLdapOUForm) (interface{}, e.Error) {
	_, err := c.DB().Where(`id = ?`, form.Id).Delete(&models.LdapOUProject{})
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	return resps.DeleteLdapOUResp{
		Id: form.Id,
	}, nil
}

func AuthProjectLdapOU(c *ctx.ServiceContext, form *forms.AuthProjectLdapOUForm) (interface{}, e.Error) {
	result, err := services.CreateOUProject(c.DB(), models.LdapOUProject{
		OrgId:     c.OrgId,
		ProjectId: models.Id(form.ProjectId),
		DN:        form.DN,
		OU:        form.OU,
		Role:      form.Role,
	})
	return result, err
}
