// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// UpdatePolicySuppress 更新策略屏蔽
// @Tags 合规/策略屏蔽
// @Summary 更新策略屏蔽
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param json body forms.UpdatePolicySuppressForm true "parameter"
// @Param policyId path string true "策略id"
// @Router /policies/{policyId}/suppress [post]
// @Success 200 {object} ctx.JSONResult{result=models.PolicySuppress}
func (Policy) UpdatePolicySuppress(c *ctx.GinRequest) {
	form := &forms.UpdatePolicySuppressForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdatePolicySuppress(c.Service(), form))
}

// DeletePolicySuppress 删除策略屏蔽
// @Tags 合规/策略屏蔽
// @Summary 删除策略屏蔽
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param policyId path string true "策略id"
// @Param suppressId path string true "屏蔽策略id"
// @Router /policies/{policyId}/suppress/{suppressId} [delete]
// @Success 200 {object} ctx.JSONResult
func (Policy) DeletePolicySuppress(c *ctx.GinRequest) {
	form := &forms.DeletePolicySuppressForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeletePolicySuppress(c.Service(), form))
}

// SearchPolicySuppress 获取策略屏蔽列表
// @Tags 合规/策略屏蔽
// @Summary 获取策略屏蔽列表。
// @Description 获取策略屏蔽列表。该列表仅返回手动设置的策略屏蔽，不包含策略组屏蔽和环境/云模板禁用扫描导致的策略屏蔽。
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param policyId path string true "策略id"
// @Param form query forms.SearchPolicySuppressForm true "parameter"
// @Router /policies/{policyId}/suppress [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.PolicySuppressResp}}
func (Policy) SearchPolicySuppress(c *ctx.GinRequest) {
	form := &forms.SearchPolicySuppressForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicySuppress(c.Service(), form))
}

// SearchPolicySuppressSource 获取策略屏蔽源列表
// @Tags 合规/策略屏蔽
// @Summary 获取策略屏蔽列表。
// @Description 获取策略屏蔽列表。该列表仅返回手动设置的策略屏蔽，
// @Description 不包含策略组屏蔽和环境/云模板禁用扫描导致的策略屏蔽。
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param policyId path string true "策略id"
// @Param form query forms.SearchPolicySuppressSourceForm true "parameter"
// @Router /policies/{policyId}/suppress/sources [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.PolicySuppressSourceResp}}
func (Policy) SearchPolicySuppressSource(c *ctx.GinRequest) {
	form := &forms.SearchPolicySuppressSourceForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicySuppressSource(c.Service(), form))
}
