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
// @Summary 创建策略组
// @Tags 合规/策略组
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param json body forms.CreatePolicyGroupForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=models.PolicyGroup}
// @Router /policies/groups [post]
func (PolicyGroup) Create(c *ctx.GinRequest) {
	form := &forms.CreatePolicyGroupForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreatePolicyGroup(c.Service(), form))
}

// Search 查询策略组列表
// @Tags 合规/策略组
// @Summary 查询策略组列表
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param q query string false "模糊搜索"
// @Router /policies/groups [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.PolicyGroup}}
func (PolicyGroup) Search(c *ctx.GinRequest) {
	form := &forms.SearchPolicyGroupForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicyGroup(c.Service(), form))
}

// Update 修改策略组
// @Tags 合规/策略组
// @Summary 修改策略组
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param json body forms.UpdatePolicyGroupForm true "parameter"
// @Param policyId path string true "策略组Id"
// @Router /policies/groups/{policyId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.PolicyGroup}
func (PolicyGroup) Update(c *ctx.GinRequest) {
	form := &forms.UpdatePolicyGroupForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdatePolicyGroup(c.Service(), form))
}

// Delete 删除策略组
// @Tags 合规/策略组
// @Summary 删除策略组
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param policyId path string true "策略组Id"
// @Router /policies/groups/{policyId} [delete]
// @Success 200 {object} ctx.JSONResult
func (PolicyGroup) Delete(c *ctx.GinRequest) {
	form := &forms.DeletePolicyGroupForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeletePolicyGroup(c.Service(), form))
}

// Detail 策略组详情
// @Tags 合规/策略组
// @Summary 策略组详情
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param policyId path string true "策略组Id"
// @Router /policies/groups/{policyId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.PolicyGroup}
func (PolicyGroup) Detail(c *ctx.GinRequest) {
	form := &forms.DetailPolicyGroupForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailPolicyGroup(c.Service(), form))
}

// OpPolicyAndPolicyGroupRel 添加/移除策略与策略组的关系
// @Tags 合规/策略组
// @Summary 添加/移除策略与策略组的关系
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param policyId path string true "策略组Id"
// @Router /policies/groups/{policyId} [post]
// @Success 200 {object} ctx.JSONResult
func (PolicyGroup) OpPolicyAndPolicyGroupRel(c *ctx.GinRequest) {
	form := &forms.OpnPolicyAndPolicyGroupRelForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.OpPolicyAndPolicyGroupRel(c.Service(), form))
}

// ScanReport 策略组详情-报表
// @Tags 合规/策略组
// @Summary 策略详情-报表
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param policyGroupId path string true "策略id"
// @Router /policies/groups/{policyGroupId}/report [get]
// @Success 200 {object} ctx.JSONResult{result=apps.PolicyGroupScanReportResp}
func (PolicyGroup) ScanReport(c *ctx.GinRequest) {
	form := &forms.PolicyScanReportForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.PolicyGroupScanReport(c.Service(), form))
}

// LastTasks 策略组最近扫描内容
// @Tags 合规/策略组
// @Summary 策略组最近扫描内容
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param policyGroupId path string true "策略id"
// @Router /policies/groups/{policyGroupId}/last_tasks [get]
// @Success 200 {object} ctx.JSONResult{result=[]apps.LastScanTaskResp}
func (PolicyGroup) LastTasks(c *ctx.GinRequest) {
	form := &forms.PolicyLastTasksForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.PolicyGroupScanTasks(c.Service(), form))
}
