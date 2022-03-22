// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

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

// Search 查询策略列表
// @Tags 合规/策略
// @Summary 查询策略列表
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param q query string false "模糊搜索"
// @Param severity query string false "严重性"
// @Param groupId query string false "策略组Id"
// @Param IaC-Org-Id header string true "组织ID"
// @Router /policies [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.PolicyResp}}
func (Policy) Search(c *ctx.GinRequest) {
	form := &forms.SearchPolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicy(c.Service(), form))
}

// Detail 策略详情
// @Tags 合规/策略
// @Summary 策略详情
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param policyId path string true "策略Id"
// @Router /policies/{policyId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.Policy}
func (Policy) Detail(c *ctx.GinRequest) {
	form := &forms.DetailPolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailPolicy(c.Service(), form))
}

// PolicyError 策略详情-错误
// @Tags 合规/策略
// @Summary 策略详情-错误
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param policyId path string true "策略id"
// @Param IaC-Org-Id header string true "组织ID"
// @Router /policies/{policyId}/error [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.PolicyErrorResp}}
func (Policy) PolicyError(c *ctx.GinRequest) {
	form := &forms.PolicyErrorForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.PolicyError(c.Service(), form))
}

// PolicyReport 策略详情-报表
// @Tags 合规/策略
// @Summary 策略详情-报表
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param policyId path string true "策略id"
// @Param IaC-Org-Id header string true "组织ID"
// @Router /policies/{policyId}/report [get]
// @Success 200 {object} ctx.JSONResult{result=resps.PolicyScanReportResp}
func (Policy) PolicyReport(c *ctx.GinRequest) {
	form := &forms.PolicyScanReportForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.PolicyScanReport(c.Service(), form))
}

// Parse 云模板/环境源码解析
// @Summary 云模板/环境源码解析
// @Description 运行云模板/环境源码解析，该 API 执行速度较慢，需要 5 ～ 15 秒，前端应明显提醒用户
// @Tags 合规/策略
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param json body forms.PolicyParseForm true "parameter"
// @Param IaC-Org-Id header string true "组织ID"
// @Success 200 {object}  ctx.JSONResult{result=resps.ParseResp}
// @Router /policies/parse [post]
func (Policy) Parse(c *ctx.GinRequest) {
	form := &forms.PolicyParseForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.ParseTemplate(c.Service(), form))
}

// Test 策略测试
// @Summary 策略测试
// @Tags 合规/策略
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param json body forms.PolicyTestForm true "parameter"
// @Param IaC-Org-Id header string true "组织ID"
// @Success 200 {object}  ctx.JSONResult{result=resps.PolicyTestResp}
// @Router /policies/test [post]
func (Policy) Test(c *ctx.GinRequest) {
	form := &forms.PolicyTestForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.PolicyTest(c.Service(), form))
}

// PolicySummary 策略概览
// @Tags 合规/策略
// @Summary 策略概览
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Router /policies/summary [get]
// @Success 200 {object} ctx.JSONResult{result=resps.PolicySummaryResp}
func (Policy) PolicySummary(c *ctx.GinRequest) {
	c.JSONResult(apps.PolicySummary(c.Service()))
}
