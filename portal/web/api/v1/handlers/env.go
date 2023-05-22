// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
)

type  Env struct {
	ctrl.GinController
}

// Create 创建环境
// @Tags 环境
// @Summary 创建环境
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param json body forms.CreateEnvForm true "环境参数"
// @router /envs [post]
// @Success 200 {object} ctx.JSONResult{result=models.EnvDetail}
func (Env) Create(c *ctx.GinRequest) {
	form := forms.CreateEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.CreateEnv(c.Service(), &form))
}

// Search 环境查询
// @Tags 环境
// @Summary 环境查询
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param form query forms.SearchEnvForm true "parameter"
// @router /envs [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.EnvDetail}}
func (Env) Search(c *ctx.GinRequest) {
	form := forms.SearchEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchEnv(c.Service(), &form))
}

// Update 环境编辑
// @Tags 环境
// @Summary 环境信息编辑
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param form body forms.UpdateEnvForm true "parameter"
// @Param envId path string true "环境ID"
// @router /envs/{envId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.EnvDetail}
func (Env) Update(c *ctx.GinRequest) { //nolint:dupl
	form := forms.UpdateEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateEnv(c.Service(), &form))
}

// Detail 环境信息详情
// @Tags 环境
// @Summary 环境信息详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @router /envs/{envId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.EnvDetail}
func (Env) Detail(c *ctx.GinRequest) {
	form := forms.DetailEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.EnvDetail(c.Service(), form))
}

// Archive 环境归档
// @Tags 环境
// @Summary 环境归档
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param form body forms.ArchiveEnvForm true "parameter"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/archive [put]
// @Success 200 {object} ctx.JSONResult{result=models.EnvDetail}
func (Env) Archive(c *ctx.GinRequest) { //nolint:dupl
	form := forms.UpdateEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateEnv(c.Service(), &form))
}

// Deploy 环境重新部署
// @Tags 环境
// @Summary 环境重新部署
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param data body forms.DeployEnvForm true "部署参数"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/deploy [post]
// @Success 200 {object} ctx.JSONResult{result=models.EnvDetail}
func (Env) Deploy(c *ctx.GinRequest) {
	form := forms.DeployEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.EnvDeploy(c.Service(), &form))
}

// DeployCheck 环境重新部署检测接口
// @Tags 环境
// @Summary 环境重新部署检测
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/deploy/check [post]
// @Success 200 {object} ctx.JSONResult{}
func (Env) DeployCheck(c *ctx.GinRequest) {
	c.JSONResult(apps.EnvDeployCheck(c.Service(), models.Id(c.Param("id"))))
}

// Destroy 销毁环境资源
// @Tags 环境
// @Summary 销毁环境资源
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param data body forms.DestroyEnvForm true "销毁参数"
// @router /envs/{envId}/destroy [post]
// @Success 200 {object} ctx.JSONResult{result=models.EnvDetail}
func (Env) Destroy(c *ctx.GinRequest) {
	form := forms.DestroyEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}

	deployForm := forms.DeployEnvForm{
		Id:       form.Id,
		TaskType: models.TaskTypeDestroy,
		Source:   form.Source,
	}
	c.JSONResult(apps.EnvDeploy(c.Service(), &deployForm))
}

// SearchResources 获取环境资源列表
// @Tags 环境
// @Summary 获取环境资源列表
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param form query forms.SearchEnvResourceForm true "parameter"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/resources [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]services.Resource}}
func (Env) SearchResources(c *ctx.GinRequest) {
	form := forms.SearchEnvResourceForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchEnvResources(c.Service(), &form))
}

// Output 环境的 Terraform Output
// @Tags 环境
// @Summary 环境的 Terraform Output
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/output [get]
// @Success 200 {object} ctx.JSONResult
func (Env) Output(c *ctx.GinRequest) {
	form := forms.DetailEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.EnvOutput(c.Service(), form))
}

// Variables 查询环境部署时使用的变量
// @Tags 环境
// @Summary 查询环境部署时使用的变量
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.SearchEnvVariableForm true "parameter"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/variables [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.VariableBody}}
func (Env) Variables(c *ctx.GinRequest) {
	form := forms.SearchEnvVariableForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.EnvVariables(c.Service(), form))
}

// SearchTasks 部署历史
// @Tags 环境
// @Summary 部署历史
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param form query forms.SearchEnvTasksForm true "parameter"
// @router /envs/{envId}/tasks [get]
// @Success 200 {object} ctx.JSONResult{result=resps.TaskDetailResp}
func (Env) SearchTasks(c *ctx.GinRequest) {
	form := &forms.SearchEnvTasksForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	taskForm := &forms.SearchTaskForm{
		NoPageSizeForm: form.NoPageSizeForm,
		EnvId:          form.Id,
		TaskType:       form.TaskType,
		Source:         form.Source,
		User:           form.User,
	}
	c.JSONResult(apps.SearchTask(c.Service(), taskForm))
}

// LastTask 环境最新任务详情
// @Tags 环境
// @Summary 环境最新任务详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/tasks/last [get]
// @Success 200 {object} ctx.JSONResult{result=resps.TaskDetailResp}
func (Env) LastTask(c *ctx.GinRequest) {
	form := &forms.LastTaskForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.LastTask(c.Service(), form))
}

// PolicyResult 环境合规详情
// @Tags 环境
// @Summary 环境合规详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param form query forms.PolicyScanResultForm true "parameter"
// @router /envs/{envId}/policy_result [get]
// @Success 200 {object} ctx.JSONResult{result=resps.ScanResultPageResp}
func (Env) PolicyResult(c *ctx.GinRequest) {
	form := &forms.PolicyScanResultForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.PolicyScanResult(c.Service(), consts.ScopeEnv, form))
}

// ResourceDetail 资源部署成功后信息详情
// @Tags 环境
// @Summary 环境部署资源信息详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param envId path string true "环境ID"
// @Param resourceId path string true "资源ID"
// @route /envs/{envId}/resources/{resourceId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.ResAttrs}
func (Env) ResourceDetail(c *ctx.GinRequest) {
	form := &forms.ResourceDetailForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.ResourceDetail(c.Service(), form))
}

// SearchResourcesGraph 获取环境资源列表
// @Tags 环境
// @Summary 获取环境资源列表
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param form query forms.SearchEnvResourceGraphForm true "parameter"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/resources/graph [get]
// @Success 200 {object} ctx.JSONResult
func (Env) SearchResourcesGraph(c *ctx.GinRequest) {
	form := forms.SearchEnvResourceGraphForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchEnvResourcesGraph(c.Service(), &form))
}

// ResourceGraphDetail 获取环境资源详情
// @Tags 环境
// @Summary 获取环境资源详情
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param resourceId path string true "资源ID"
// @router /envs/{envId}/resources/graph/{resourceId} [get]
// @Success 200 {object} ctx.JSONResult{result=services.Resource}
func (Env) ResourceGraphDetail(c *ctx.GinRequest) {
	form := &forms.ResourceGraphDetailForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.ResourceGraphDetail(c.Service(), form))
}

// UpdateTags 全量更新环境 Tags
// @Tags 环境
// @Summary 全量更新环境 Tags
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param tags formData forms.UpdateEnvTagsForm true "部署参数"
// @router /envs/{envId}/tags [post]
// @Success 200 {object} ctx.JSONResult{result=models.Env}
func (Env) UpdateTags(c *ctx.GinRequest) {
	form := forms.UpdateEnvTagsForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.EnvUpdateTags(c.Service(), &form))
}

// EnvStat 环境概览
// @Tags 环境
// @Summary 环境概览
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @router /envs/{Id}/statistics [get]
// @Success 200 {object} ctx.JSONResult{result=resps.EnvStatisticsResp}
func (Env) EnvStat(c *ctx.GinRequest) {
	form := forms.EnvParam{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.EnvStat(c.Service(), &form))
}

// EnvLock 环境锁定
// @Tags 环境
// @Summary 环境锁定
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param tags formData forms.EnvLockForm true "部署参数"
// @router /envs/{envId}/lock [post]
// @Success 200
func EnvLock(c *ctx.GinRequest) {
	form := forms.EnvLockForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.EnvLock(c.Service(), &form))
}

// EnvUnLock 环境解锁
// @Tags 环境
// @Summary 环境解锁
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param tags formData forms.EnvUnLockForm true "部署参数"
// @router /envs/{envId}/unlock [post]
// @Success 200
func EnvUnLock(c *ctx.GinRequest) {
	form := forms.EnvUnLockForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.EnvUnLock(c.Service(), &form))
}

// EnvUnLockConfirm 环境解锁确认
// @Tags 环境
// @Summary 环境解锁确认
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param tags formData forms.EnvUnLockConfirmForm true "部署参数"
// @router /envs/{envId}/unlock/confirm [get]
// @Success 200 {object} ctx.JSONResult{result=resps.EnvUnLockConfirmResp}
func EnvUnLockConfirm(c *ctx.GinRequest) {
	form := forms.EnvUnLockConfirmForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.EnvUnLockConfirm(c.Service(), &form))
}
