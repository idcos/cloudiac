// Copyright 2021 CloudJ Company Limited. All rights reserved.

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

// Update 修改策略组信息
// @Tags 合规/策略组
// @Summary 修改策略组信息
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param json body forms.UpdatePolicyGroupForm true "parameter"
// @Param policyGroupId path string true "策略组Id"
// @Router /policies/groups/{policyGroupId} [put]
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
// @Param policyGroupId path string true "策略组Id"
// @Router /policies/groups/{policyGroupId} [delete]
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
// @Param policyGroupId path string true "策略组Id"
// @Router /policies/groups/{policyGroupId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.PolicyGroup}
func (PolicyGroup) Detail(c *ctx.GinRequest) {
	form := &forms.DetailPolicyGroupForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailPolicyGroup(c.Service(), form))
}

// SearchGroupOfPolicy 查询策略组关联的策略或未关联策略组的策略
// @Tags 合规/策略组
// @Summary 查询策略组关联的策略或未关联策略组的策略
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param bind query bool false "是否查询绑定了策略组的策略 ture: 查询绑定策略组的策略，false: 查询未绑定的策略组的策略"
// @Param policyGroupId path string true "策略组id"
// @Router /policies/groups/{policyGroupId}/policies [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Policy}}
func (PolicyGroup) SearchGroupOfPolicy(c *ctx.GinRequest) {
	form := &forms.SearchGroupOfPolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchGroupOfPolicy(c.Service(), form))
}

// OpPolicyAndPolicyGroupRel 添加/移除策略与策略组的关系
// @Tags 合规/策略组
// @Summary 添加/移除策略与策略组的关系
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param json body forms.OpnPolicyAndPolicyGroupRelForm true "parameter"
// @Param policyGroupId path string true "策略组Id"
// @Router /policies/groups/{policyGroupId} [post]
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

// PolicyGroupChecks
// @Tags 合规/策略组
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Summary 创建策略组前检查名称工作目录是否正确
// @Param IaC-Org-Id header string true "组织ID"
// @Security AuthToken
// @router /policies/groups/checks [POST]
// @Param form query forms.PolicyGroupChecksForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=apps.TemplateChecksResp}
func PolicyGroupChecks(c *ctx.GinRequest) {
	form := forms.PolicyGroupChecksForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.TemplateChecks(c.Service(), &forms.TemplateChecksForm{
		Name:         "",
		RepoId:       form.RepoId,
		RepoRevision: form.RepoRevision,
		VcsId:        form.VcsId,
		Workdir:      form.Dir,
	}))
}

// SearchRegistryPG registry侧策略组列表
// @Tags policy_group
// @Summary registry侧策略组列表
// @Accept application/x-www-form-urlencoded, application/json
// @Produce json
// @Param json body forms.SearchRegistryPgForm true "parameter"
// @router /vcs/registry/policy_groups [GET]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]apps.RegistryPgResp}}

func SearchRegistryPG(c *ctx.GinRequest) {
	form := forms.SearchRegistryPgForm{}
	if err := c.Bind(&form); err != nil {
		return
	}

	c.JSONResult(apps.SearchRegistryPG(c.Service(), &form))
}

// SearchRegistryPGVersions registry侧策略组版本列表
// @Tags policy_group
// @Summary registry侧策略组版本列表
// @Accept application/x-www-form-urlencoded, application/json
// @Produce json
// @Param json body forms.SearchRegistryPgVersForm true "parameter"
// @router /vcs/registry/policy_groups/versions [GET]
// @Success 200 {object} ctx.JSONResult{result=[]apps.RegistryPGVerResp}
func SearchRegistryPGVersions(c *ctx.GinRequest) {
	form := forms.SearchRegistryPgVersForm{}
	if err := c.Bind(&form); err != nil {
		return
	}

	c.JSONResult(apps.SearchRegistryPGVersions(c.Service(), &form))
}
