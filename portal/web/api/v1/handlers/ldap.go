// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// GetLdapOUs ldap ou 列表
// @Summary ldap ou 列表
// @Tags ldap
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Success 200 {object}  ctx.JSONResult{result=resps.LdapOUListResp}
// @Router /orgs/{id}/ldap_ous [get]
func GetLdapOUs(c *ctx.GinRequest) {
	c.JSONResult(apps.GetLdapOUs(c.Service()))
}

// GetLdapUsers ldap user 列表
// @Summary ldap user 列表
// @Tags ldap
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param form query forms.SearchLdapUserForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=resps.LdapUserListResp}
// @Router /orgs/{id}/ldap_users [get]
func GetLdapUsers(c *ctx.GinRequest) {
	form := &forms.SearchLdapUserForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.GetLdapUsers(c.Service(), form))
}
