package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Policy struct {
	ctrl.GinController
}

// Create 创建策略
// @Summary 创建策略
// @Description 创建策略
// @Tags 策略
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param json body forms.CreatePolicyForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=models.Policy}
// @Router /policies [post]
func (Policy) Create(c *ctx.GinRequest) {
	form := &forms.CreatePolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreatePolicy(c.Service(), form))
}

// Parse 解析云模板
// @Summary 解析云模板
// @Description 解析云模板
// @Tags 策略
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param json body forms.CreatePolicyForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=models.Policy}
// @Router /policies [post]
func (Policy) Parse(c *ctx.GinRequest) {
	form := &forms.CreatePolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreatePolicy(c.Service(), form))
}

// Scan 运行策略扫描
// @Summary 运行策略扫描
// @Description 运行策略扫描
// @Tags 策略
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param json body forms.CreatePolicyForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=models.Policy}
// @Router /policies [post]
func (Policy) Scan(c *ctx.GinRequest) {
	form := &forms.CreatePolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreatePolicy(c.Service(), form))
}
