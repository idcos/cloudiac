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
	c.JSONResult(apps.ScanTemplate(c.Service(), form, ""))
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
// @Router /policies/template/{templateId}/result [get]
// @Success 200 {object} ctx.JSONResult{result=apps.ScanResultResp}
func (Policy) TemplateScanResult(c *ctx.GinRequest) {
	form := &forms.PolicyScanResultForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	form.Scope = consts.ScopeTemplate
	c.JSONResult(apps.PolicyScanResult(c.Service(), form))
}

// SearchPolicyTpl 查询云模板策略配置
// @Tags 合规/云模板
// @Summary 查询云模板策略配置
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form query forms.SearchPolicyTplForm true "parameter"
// @Param IaC-Org-Id header string true "组织ID"
// @Router /policies/templates [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]apps.RespPolicyTpl}}
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
// @Param json body forms.UpdatePolicyRelForm true "parameter"
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Router /policies/templates/{templateId} [put]
// @Success 200 {object} ctx.JSONResult
func (Policy) UpdatePolicyTpl(c *ctx.GinRequest) {
	form := &forms.UpdatePolicyRelForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	form.Scope = consts.ScopeTemplate
	c.JSONResult(apps.UpdatePolicyRel(c.Service(), form))
}
