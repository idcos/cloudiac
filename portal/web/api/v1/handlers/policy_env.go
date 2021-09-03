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
// @Param q query string false "模糊搜索"
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Router /policies/envs [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]apps.RespPolicyEnv}}
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
// @Param IaC-Project-Id header string false "项目ID"
// @Param envId path string true "环境id"
// @Router /policies/envs/{envId}/policies [get]
// @Success 200 {object} ctx.JSONResult{result=models.Policy}
func (Policy) EnvOfPolicy(c *ctx.GinRequest) {
	form := &forms.EnvOfPolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.EnvOfPolicy(c.Service(), form))
}

// UpdatePolicyEnv 修改环境与策略组关联
// @Tags 合规/环境
// @Summary 修改环境与策略组关联
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param json body forms.UpdatePolicyEnvForm true "parameter"
// @Router /policies/envs [put]
// @Success 200 {object} ctx.JSONResult
func (Policy) UpdatePolicyEnv(c *ctx.GinRequest) {
	//form := &forms.UpdatePolicyEnvForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.UpdatePolicyEnv(c.Service(), form))
}

// ScanEnvironment 运行环境策略扫描
// @Summary 运行环境策略扫描
// @Tags 合规/环境
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
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
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.PolicyScanResultForm true "parameter"
// @Param envId path string true "环境ID"
// @Router /policies/envs/{envId}/result [get]
// @Success 200 {object} ctx.JSONResult{result=apps.ScanResultResp}
func (Policy) EnvScanResult(c *ctx.GinRequest) {
	form := &forms.PolicyScanResultForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	form.Scope = consts.ScopeEnv
	c.JSONResult(apps.PolicyScanResult(c.Service(), form))
}
