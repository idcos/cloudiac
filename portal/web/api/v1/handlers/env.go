// Copyright 2021 CloudJ Company Limited. All rights reserved.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
)

type Env struct {
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
// @Success 200 {object} ctx.JSONResult{result=models.Env}
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
// @Param form body forms.ArchiveEnvForm true "parameter"
// @Param envId path string true "环境ID"
// @router /envs/{envId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.Env}
func (Env) Update(c *ctx.GinRequest) {
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
func (Env) Archive(c *ctx.GinRequest) {
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
// @router /envs/{envId}/destroy [post]
// @Success 200 {object} ctx.JSONResult{result=models.EnvDetail}
func (Env) Destroy(c *ctx.GinRequest) {
	form := forms.DeployEnvForm{}
	form.Id = models.Id(c.Param("id"))
	form.TaskType = models.TaskTypeDestroy
	c.JSONResult(apps.EnvDeploy(c.Service(), &form))
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
// @Success 200 {object} ctx.JSONResult{result=models.Resource}
func (Env) SearchResources(c *ctx.GinRequest) {
	form := forms.SearchEnvResourceForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchEnvResources(c.Service(), &form))
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
// @Success 200 {object} ctx.JSONResult{result=apps.taskDetailResp}
func (Env) SearchTasks(c *ctx.GinRequest) {
	form := &forms.SearchEnvTasksForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	f := &forms.SearchTaskForm{EnvId: form.Id}
	c.JSONResult(apps.SearchTask(c.Service(), f))
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
// @Success 200 {object} ctx.JSONResult{result=apps.taskDetailResp}
func (Env) LastTask(c *ctx.GinRequest) {
	form := &forms.LastTaskForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.LastTask(c.Service(), form))
}

// ResourceDetail 资源部署成功后信息详情
// @Tags 环境
// @Summary 环境部署资源信息详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param resourceId path string true "资源ID"
// @route /envs/{envId}/resource/{resourceId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.ResAttrs}
func (Env) ResourceDetail(c *ctx.GinRequest) {
	form := &forms.ResourceDetailForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.ResourceDetail(c.Service(), form))
}
