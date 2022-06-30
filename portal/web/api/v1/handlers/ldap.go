// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// GetLdapOUsFromLdap ldap ou 列表(from ldap)
// @Summary ldap ou 列表(from ldap)
// @Tags ldap
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Success 200 {object}  ctx.JSONResult{result=resps.LdapOUResp}
// @Router /ldap/ous [get]
func GetLdapOUsFromLdap(c *ctx.GinRequest) {
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
// @Router /ldap/users [get]
func GetLdapUsers(c *ctx.GinRequest) {
	form := &forms.SearchLdapUserForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.GetLdapUsers(c.Service(), form))
}

// AuthLdapUser ldap user 授权
// @Summary ldap user 授权
// @Tags ldap
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param form query forms.AuthLdapUserForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=resps.AuthLdapUserResp}
// @Router /ldap/auth/org_user [post]
func AuthLdapUser(c *ctx.GinRequest) {
	form := &forms.AuthLdapUserForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.AuthLdapUser(c.Service(), form))
}

// AuthLdapOU ldap ou 授权
// @Summary ldap ou 授权
// @Tags ldap
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param form query forms.AuthLdapOUForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=resps.AuthLdapOUResp}
// @Router /ldap/auth/org_ou [post]
func AuthLdapOU(c *ctx.GinRequest) {
	form := &forms.AuthLdapOUForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.AuthLdapOU(c.Service(), form))
}

// GetLdapOUsFromDB ldap ou 列表
// @Summary ldap ou 列表
// @Tags ldap
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param form query forms.SearchLdapOUForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.LdapOUDBResp}}
// @Router /ldap/org_ous [get]
func GetLdapOUsFromDB(c *ctx.GinRequest) {
	form := &forms.SearchLdapOUForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.GetLdapOUsFromDB(c.Service(), form))
}

// DeleteLdapOUFromDB 删除 ldap ou
// @Summary 删除 ldap ou
// @Tags ldap
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param form query forms.DeleteLdapOUForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=resps.DeleteLdapOUResp}
// @Router /ldap/org_ou [delete]
func DeleteLdapOUFromDB(c *ctx.GinRequest) {
	form := &forms.DeleteLdapOUForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteLdapOUFromDB(c.Service(), form))
}

// UpdateLdapOU 更新 org 的 ldap ou
// @Summary 更新 org 的 ldap ou
// @Tags ldap
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param form query forms.UpdateLdapOUForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=resps.UpdateLdapOUResp}
// @Router /ldap/project_ou [PUT]
func UpdateLdapOU(c *ctx.GinRequest) {
	form := &forms.UpdateLdapOUForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateLdapOU(c.Service(), form))
}

// GetLdapOUsFromOrg 获取组织下的 ldap ou 列表
// @Summary 获取组织下的 ldap ou 列表
// @Tags ldap
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param form query forms.SearchLdapOUForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.LdapOUDBResp}}
// @Router /ldap/project_ous [get]
func GetProjectLdapOUs(c *ctx.GinRequest) {
	form := &forms.SearchLdapOUForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.GetOrgLdapOUs(c.Service(), form))
}

// DeleteProjectLdapOU 移除 project的 ldap ou
// @Summary 移除 project的 ldap ou
// @Tags ldap
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param form query forms.DeleteLdapOUForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=resps.DeleteLdapOUResp}
// @Router /ldap/project_ou [delete]
func DeleteProjectLdapOU(c *ctx.GinRequest) {
	form := &forms.DeleteLdapOUForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteProjectLdapOU(c.Service(), form))
}

// UpdateProjectLdapOU 更新 project的 ldap ou
// @Summary 更新 project的 ldap ou
// @Tags ldap
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param form query forms.UpdateLdapOUForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=resps.UpdateLdapOUResp}
// @Router /ldap/project_ou [PUT]
func UpdateProjectLdapOU(c *ctx.GinRequest) {
	form := &forms.UpdateLdapOUForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateProjectLdapOU(c.Service(), form))
}

// AuthProjectLdapOU ldap ou 授权
// @Summary ldap ou 授权
// @Tags ldap
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param form query forms.AuthProjectLdapOUForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=resps.AuthLdapOUResp}
// @Router /ldap/auth/project_ou [post]
func AuthProjectLdapOU(c *ctx.GinRequest) {
	form := &forms.AuthProjectLdapOUForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.AuthProjectLdapOU(c.Service(), form))
}
