package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// UpdatePolicySuppress 更新策略屏蔽
// @Tags 合规/策略
// @Summary 更新策略屏蔽
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param policyId path string true "策略id"
// @Router /policies/{policyId}/suppress [post]
// @Success 200 {object} ctx.JSONResult
func (Policy) UpdatePolicySuppress(c *ctx.GinRequest) {
	form := &forms.UpdatePolicySuppressForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdatePolicySuppress(c.Service(), form))
}

// DeletePolicySuppress 删除策略屏蔽
// @Tags 合规/策略
// @Summary 删除策略屏蔽
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
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
// @Tags 合规/策略
// @Summary 获取策略屏蔽列表。
// @Description 获取策略屏蔽列表。该列表仅返回手动设置的策略屏蔽，不包含策略组屏蔽和环境/云模板禁用扫描导致的策略屏蔽。
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param policyId path string true "策略id"
// @Router /policies/{policyId}/suppress [get]
// @Success 200 {object} ctx.JSONResult{result=[]apps.PolicySuppressResp}
func (Policy) SearchPolicySuppress(c *ctx.GinRequest) {
	form := &forms.SearchPolicySuppressForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicySuppress(c.Service(), form))
}

// SearchPolicySuppressSource 获取策略屏蔽源列表
// @Tags 合规/策略
// @Summary 获取策略屏蔽列表。
// @Description 获取策略屏蔽列表。该列表仅返回手动设置的策略屏蔽，不包含策略组屏蔽和环境/云模板禁用扫描导致的策略屏蔽。该列表自动过滤器已经屏蔽的源
// @Accept json
// @Produce json
// @Security AuthToken
// @Param policyId path string true "策略id"
// @Router /policies/{policyId}/suppress/sources [get]
// @Success 200 {object} ctx.JSONResult{result=[]apps.PolicySuppressSourceResp}
func (Policy) SearchPolicySuppressSource(c *ctx.GinRequest) {
	form := &forms.SearchPolicySuppressSourceForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicySuppressSource(c.Service(), form))
}
