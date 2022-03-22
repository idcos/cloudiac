// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

// ScanTemplate 运行云模板策略扫描
// @Summary 运行云模板策略扫描
// @Tags 合规/云模板
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param templateId path string true "模板ID"
// @Param json body forms.ScanTemplateForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=models.ScanTask}
// @Router /policies/templates/{templateId}/scan [post]
func (Policy) ScanTemplate(c *ctx.GinRequest) {
	form := &forms.ScanTemplateForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.ScanTemplateOrEnv(c.Service(), form, ""))
}

// ScanTemplates 运行多个云模板策略扫描
// @Summary 运行云模板策略扫描
// @Tags 合规/云模板
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param json body forms.ScanTemplateForms true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=[]models.ScanTask}
// @Router /policies/templates/scans [post]
func (Policy) ScanTemplates(c *ctx.GinRequest) {
	form := &forms.ScanTemplateForms{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.ScanTemplates(c.Service(), form))

}

// TemplateScanResult 云模板策略扫描结果
// @Tags 合规/云模板
// @Summary 云模板策略扫描结果
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.PolicyScanResultForm true "parameter"
// @Param templateId path string true "环境ID"
// @Router /policies/templates/{templateId}/result [get]
// @Success 200 {object} ctx.JSONResult{result=resps.ScanResultPageResp}
func (Policy) TemplateScanResult(c *ctx.GinRequest) {
	form := &forms.PolicyScanResultForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.PolicyScanResult(c.Service(), consts.ScopeTemplate, form))
}

// SearchPolicyTpl 查询云模板策略配置
// @Tags 合规/云模板
// @Summary 查询云模板策略配置
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.SearchPolicyTplForm true "parameter"
// @Router /policies/templates [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.RespPolicyTpl}}
func (Policy) SearchPolicyTpl(c *ctx.GinRequest) {
	form := &forms.SearchPolicyTplForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicyTpl(c.Service(), form))
}

// UpdatePolicyTpl 修改云模板与策略组关联
// @Tags 合规/云模板
// @Summary 修改云模板与策略组关联
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param json body forms.UpdatePolicyRelForm true "parameter"
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Router /policies/templates/{templateId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.PolicyRel}
func (Policy) UpdatePolicyTpl(c *ctx.GinRequest) {
	form := &forms.UpdatePolicyRelForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	form.Scope = consts.ScopeTemplate
	c.JSONResult(apps.UpdatePolicyRelNew(c.Service(), form))
}

// TplOfPolicy 云模板策略详情
// @Tags 合规/云模板
// @Summary 云模板策略详情
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param templateId path string true "云模板id"
// @Router /policies/templates/{templateId}/policies [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.RespTplOfPolicy}}
func (Policy) TplOfPolicy(c *ctx.GinRequest) {
	form := &forms.TplOfPolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.TplOfPolicy(c.Service(), form))
}

// TplOfPolicyGroup 云模板策略组详情
// @Tags 合规/云模板
// @Summary 云模板策略组详情
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param templateId path string true "云模板id"
// @Router /policies/templates/{templateId}/policies/groups [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.PolicyGroup}}
func (Policy) TplOfPolicyGroup(c *ctx.GinRequest) {
	form := &forms.TplOfPolicyGroupForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.TplOfPolicyGroup(c.Service(), form))
}

//todo swagger文档缺失
func (Policy) ValidTplOfPolicy(c *ctx.GinRequest) {
	form := &forms.TplOfPolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.ValidTplOfPolicy(c.Service(), form))
}

// EnablePolicyTpl 启用云模板扫描
// @Tags 合规/云模板
// @Summary 启用云模板扫描
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param json body forms.EnableScanForm true "parameter"
// @Param templateId path string true "云模板ID"
// @Router /policies/templates/{templateId}/enabled [put]
// @Success 200 {object} ctx.JSONResult
func (Policy) EnablePolicyTpl(c *ctx.GinRequest) {
	form := &forms.EnableScanForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	form.Scope = consts.ScopeTemplate
	c.JSONResult(apps.EnablePolicyScanRel(c.Service(), form))
}
