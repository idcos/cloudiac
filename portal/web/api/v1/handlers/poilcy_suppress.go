package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type PolicySuppress struct {
	ctrl.GinController
}

// CreatePolicySuppress 创建策略屏蔽
// @Tags 策略
// @Summary shield
// @Description 创建策略屏蔽
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param json body forms.CreatePolicyShieldForm true "parameter"
// @Router /policies/suppress [post]
// @Success 200 {object} ctx.JSONResult{result=models.PolicyShield}
func (PolicySuppress) CreatePolicySuppress(c *ctx.GinRequest) {
	form := &forms.CreatePolicyShieldForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreatePolicySuppress(c.Service(), form))
}

// SearchPolicySuppress 查询策略屏蔽
// @Tags 策略
// @Summary shield
// @Description 查询策略屏蔽
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param json body forms.SearchPolicySuppressForm true "parameter"
// @Router /policies/suppress [get]
// @Success 200 {object} ctx.JSONResult{result=models.PolicyShield}
func (PolicySuppress) SearchPolicySuppress(c *ctx.GinRequest) {
	form := &forms.SearchPolicySuppressForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicySuppress(c.Service(), form))
}

// DeletePolicySuppress 策略屏蔽
// @Tags 策略
// @Summary shield
// @Description 策略屏蔽
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param json body forms.DeletePolicySuppressForm true "parameter"
// @Router /policies/suppress [delete]
// @Success 200 {object} ctx.JSONResult{result=models.Policy}
func (PolicySuppress) DeletePolicySuppress(c *ctx.GinRequest) {
	form := &forms.DeletePolicySuppressForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeletePolicySuppress(c.Service(), form))
}
