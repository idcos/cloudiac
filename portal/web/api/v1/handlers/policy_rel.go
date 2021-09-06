package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type PolicyRel struct {
	ctrl.GinController
}

// Create 创建策略组关系
// @Summary 创建策略组关系
// @Description 创建策略组关系
// @Tags 策略
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param json body forms.UpdatePolicyRelForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=[]models.PolicyRel}
// @Router /policies/rels [post]
func (PolicyRel) Create(c *ctx.GinRequest) {
	form := &forms.UpdatePolicyRelForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdatePolicyRel(c.Service(), form))
}
