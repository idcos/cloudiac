package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type PolicyGroup struct {
	ctrl.GinController
}

// Create 创建策略组
// @Summary 创建策略
// @Description 创建策略
// @Tags 策略
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param json body forms.CreatePolicyGroupForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=models.PolicyGroup}
// @Router /policies [post]
func (PolicyGroup) Create(c *ctx.GinRequest) {
	form := &forms.CreatePolicyGroupForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreatePolicyGroup(c.Service(), form))
}
