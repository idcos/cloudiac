// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
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
