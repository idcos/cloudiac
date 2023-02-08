// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// SearchPolicyEnv 查询环境策略配置
// @Tags 合规/环境
// @Summary 查询环境策略配置
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.SearchPolicyEnvForm true "parameter"
// @Router /policies/envs [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.RespPolicyEnv}}
func (Policy) SearchPolicyEnv(c *ctx.GinRequest) {
	form := &forms.SearchPolicyEnvForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicyEnv(c.Service(), form))
}

// EnvOfPolicy 环境策略详情
// @Tags 合规/环境
// @Summary 环境策略详情
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param envId path string true "环境id"
// @Param form query forms.EnvOfPolicyForm true "parameter"
// @Router /policies/envs/{envId}/policies [get]
// @Success 200 {object} ctx.JSONResult{result=ctx.JSONResult{list=[]resps.RespEnvOfPolicy}}
func (Policy) EnvOfPolicy(c *ctx.GinRequest) {
	form := &forms.EnvOfPolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.EnvOfPolicy(c.Service(), form))
}

// ValidEnvOfPolicy 生效的环境策略
// @Tags 合规/环境
// @Summary 生效的环境策略
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param envId path string true "环境id"
// @Param form query forms.EnvOfPolicyForm true "parameter"
// @Router /policies/envs/{envId}/valid_policies [get]
// @Success 200 {object} ctx.JSONResult{result=resps.ValidPolicyResp}
func (Policy) ValidEnvOfPolicy(c *ctx.GinRequest) {
	form := &forms.EnvOfPolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.ValidEnvOfPolicy(c.Service(), form))
}

// UpdatePolicyEnv 修改环境与策略组关联
// @Tags 合规/环境
// @Summary 修改环境与策略组关联
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param json body forms.UpdatePolicyRelForm true "parameter"
// @Param envId path string true "环境ID"
// @Router /policies/envs/{envId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.PolicyRel}
func (Policy) UpdatePolicyEnv(c *ctx.GinRequest) {
	form := &forms.UpdatePolicyRelForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	form.Scope = consts.ScopeEnv
	c.JSONResult(apps.UpdatePolicyRelNew(c.Service(), form))
}

// ScanEnvironment 运行环境策略扫描
// @Summary 运行环境策略扫描
// @Tags 合规/环境
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param envId path string true "环境ID"
// @Param json body forms.ScanEnvironmentForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=models.ScanTask}
// @Router /policies/envs/{envId}/scan [post]
func (Policy) ScanEnvironment(c *ctx.GinRequest) {
	form := &forms.ScanEnvironmentForm{}
	if err := c.Bind(form); err != nil {
		return
	}

	c.JSONResult(apps.ScanEnvironment(c.Service(), form))
}

// EnvScanResult 环境策略扫描结果
// @Tags 合规/环境
// @Summary 环境策略扫描结果
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.PolicyScanResultForm true "parameter"
// @Param envId path string true "环境ID"
// @Router /policies/envs/{envId}/result [get]
// @Success 200 {object} ctx.JSONResult{result=resps.ScanResultPageResp}
func (Policy) EnvScanResult(c *ctx.GinRequest) {
	form := &forms.PolicyScanResultForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.PolicyScanResult(c.Service(), consts.ScopeEnv, form))
}

// EnablePolicyEnv 启用环境扫描
// @Tags 合规/环境
// @Summary 启用环境扫描
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param json body forms.EnableScanForm true "parameter"
// @Param envId path string true "环境ID"
// @Router /policies/envs/{envId}/enabled [put]
// @Success 200 {object} ctx.JSONResult
func (Policy) EnablePolicyEnv(c *ctx.GinRequest) {
	form := &forms.EnableScanForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	form.Scope = consts.ScopeEnv
	c.JSONResult(apps.EnablePolicyScanRel(c.Service(), form))
}
