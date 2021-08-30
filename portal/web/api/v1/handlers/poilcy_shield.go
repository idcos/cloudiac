package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type PolicyShield struct {
	ctrl.GinController
}

// CreatePolicyShield 创建策略屏蔽
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
// @Router /policies/shield [post]
// @Success 200 {object} ctx.JSONResult{result=models.PolicyShield}
func (PolicyShield) CreatePolicyShield(c *ctx.GinRequest) {
	form := &forms.CreatePolicyShieldForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreatePolicyShield(c.Service(), form))
}

// SearchPolicyShield 查询策略屏蔽
// @Tags 策略
// @Summary shield
// @Description 查询策略屏蔽
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param json body forms.SearchPolicyShieldForm true "parameter"
// @Router /policies/shield [get]
// @Success 200 {object} ctx.JSONResult{result=models.PolicyShield}
func (PolicyShield) SearchPolicyShield(c *ctx.GinRequest) {
	form := &forms.SearchPolicyShieldForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicyShield(c.Service(), form))
}

// DeletePolicyShield 策略屏蔽
// @Tags 策略
// @Summary shield
// @Description 策略屏蔽
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param json body forms.DeletePolicyShieldForm true "parameter"
// @Router /policies/shield [delete]
// @Success 200 {object} ctx.JSONResult{result=models.Policy}
func (PolicyShield) DeletePolicyShield(c *ctx.GinRequest) {
	form := &forms.DeletePolicyShieldForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeletePolicyShield(c.Service(), form))
}
