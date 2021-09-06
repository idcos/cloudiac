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
// @Tags 合规/策略
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

// Search 查询策略列表
// @Tags 合规/策略
// @Summary 查询策略列表
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param q query string false "模糊搜索"
// @Param severity query string false "严重性"
// @Param groupId query string false "策略组Id"
// @Router /policies [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Policy}}
func (Policy) Search(c *ctx.GinRequest) {
	form := &forms.SearchPolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicy(c.Service(), form))
}

// Update 修改策略
// @Tags 合规/策略
// @Summary 修改策略
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param json body forms.UpdatePolicyForm true "parameter"
// @Param policyId path string true "策略Id"
// @Router /policies/{policyId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.Policy}
func (Policy) Update(c *ctx.GinRequest) {
	form := &forms.UpdatePolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdatePolicy(c.Service(), form))
}

// Delete 删除策略
// @Tags 合规/策略
// @Summary 删除策略
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param policyId path string true "策略Id"
// @Router /policies/{policyId} [delete]
// @Success 200 {object} ctx.JSONResult
func (Policy) Delete(c *ctx.GinRequest) {
	form := &forms.DeletePolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeletePolicy(c.Service(), form))
}

// Detail 策略详情
// @Tags 合规/策略
// @Summary 策略详情
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
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
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param policyId path string true "策略id"
// @Router /policies/{policyId}/error [get]
// @Success 200 {object} ctx.JSONResult{result=apps.PolicyErrorResp}
func (Policy) PolicyError(c *ctx.GinRequest) {
	form := &forms.PolicyErrorForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.PolicyError(c.Service(), form))
}

// UpdatePolicySuppress 更新策略屏蔽
// @Tags 合规/策略
// @Summary 更新策略屏蔽
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param policyId path string true "策略id"
// @Router /policies/{policyId}/suppress [put]
// @Success 200 {object} ctx.JSONResult
func (Policy) UpdatePolicySuppress(c *ctx.GinRequest) {
	form := &forms.UpdatePolicySuppressForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdatePolicySuppress(c.Service(), form))
}

// SearchPolicySuppress 获取策略屏蔽列表
// @Tags 合规/策略
// @Summary 获取策略屏蔽列表。该列表仅返回手动设置的策略屏蔽，不包含策略组屏蔽和环境/云模板禁用扫描导致的策略屏蔽。
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param policyId path string true "策略id"
// @Router /policies/{policyId}/suppress [get]
// @Success 200 {object} ctx.JSONResult{result=[]apps.PolicySuppressResp}
func (Policy) SearchPolicySuppress(c *ctx.GinRequest) {
	form := &forms.SearchPolicySuppressForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicySuppress(c.Service(), form))
}

// PolicyReport 策略详情-报表
// @Tags 合规/策略
// @Summary 策略详情-报表
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param policyId path string true "策略id"
// @Router /policies/{policyId}/report [get]
// @Success 200 {object} ctx.JSONResult{result=apps.PolicyScanReportResp}
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
// @Param IaC-Org-Id header string true "组织id"
// @Param json body forms.PolicyParseForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=apps.ParseResp}
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
// @Param IaC-Org-Id header string true "组织id"
// @Param json body forms.PolicyTestForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=apps.PolicyTestResp}
// @Router /policies/test [post]
func (Policy) Test(c *ctx.GinRequest) {
	form := &forms.PolicyTestForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.PolicyTest(c.Service(), form))
}

// SearchGroupOfPolicy 查询策略组关联的策略或未关联策略组的策略
// @Tags 合规/策略组
// @Summary 查询策略组关联的策略或未关联策略组的策略
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param bind query bool false "是否查询绑定了策略组的策略 ture: 查询绑定策略组的策略，false: 查询未绑定的策略组的策略"
// @Param groupId path string true "策略组id"
// @Router /policies/groups/{groupId}/policies [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Policy}}
func (Policy) SearchGroupOfPolicy(c *ctx.GinRequest) {
	form := &forms.SearchGroupOfPolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchGroupOfPolicy(c.Service(), form))
}
