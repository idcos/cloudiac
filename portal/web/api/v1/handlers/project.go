// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package handlers //nolint:dupl

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Project struct {
	ctrl.GinController
}

// Create 创建项目
// @Summary 创建项目
// @Description 创建项目
// @Tags 项目
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param json body forms.CreateProjectForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=models.Project}
// @Router /projects [post]
func (Project) Create(c *ctx.GinRequest) {
	form := &forms.CreateProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateProject(c.Service(), form))
}

// Search 查询项目列表
// @Summary 查询项目列表
// @Description 查询项目列表
// @Tags 项目
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param form query forms.SearchProjectForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.ProjectResp}}
// @Router /projects [get]
func (Project) Search(c *ctx.GinRequest) {
	form := &forms.SearchProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchProject(c.Service(), form))
}

// Update 修改项目信息
// @Summary 修改项目信息
// @Description 修改项目信息
// @Tags 项目
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param projectId path string true "项目id"
// @Param request body forms.UpdateProjectForm true "用户授权"
// @Success 200 {object} ctx.JSONResult{result=models.Project}
// @Router /projects/{projectId}  [put]
func (Project) Update(c *ctx.GinRequest) {
	form := &forms.UpdateProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateProject(c.Service(), form))
}

// Delete 删除项目信息
// @Summary 删除项目信息
// @Description 删除项目信息
// @Tags 项目
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param projectId path string true "项目id"
// @Success 200
// @Router /projects/{projectId} [delete]
func (Project) Delete(c *ctx.GinRequest) {
	form := &forms.DeleteProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteProject(c.Service(), form))
}

// Detail 查询项目详情
// @Summary 查询项目详情
// @Description 查询项目详情
// @Tags 项目
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param projectId path string true "项目id"
// @Success 200 {object} ctx.JSONResult{result=resps.DetailProjectResp}
// @Router /projects/{projectId}  [get]
func (Project) Detail(c *ctx.GinRequest) {
	form := &forms.DetailProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailProject(c.Service(), form))
}

//SearchProjectResources 搜索当前项目下所有的活跃资源列表
//@Tags 项目
//@Summary 搜索当前项目下所有的活跃资源列表
//@Accept application/x-www-form-urlencoded
//@Produce json
//@Security AuthToken
//@Param IaC-Org-Id header string true "组织ID"
//@Param IaC-Project-Id header string true "项目ID"
//@Param form query forms.SearchProjectResourceForm true "parameter"
//@router /projects/resources [get]
//@Success 200 {object} ctx.JSONResult{result=resps.OrgOrProjectResourcesResp}
func (Project) SearchProjectResources(c *ctx.GinRequest) {
	form := forms.SearchProjectResourceForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchProjectResources(c.Service(), &form))
}

//SearchProjectResourcesFilters 搜索当前项目下所有的环境和provider列表
//@Tags 项目
//@Summary 搜索当前项目下所有的环境和provider列表
//@Accept application/x-www-form-urlencoded
//@Produce json
//@Security AuthToken
//@Param IaC-Org-Id header string true "组织ID"
//@Param IaC-Project-Id header string true "项目ID"
//@Param form query forms.SearchProjectResourceForm true "parameter"
//@router /projects/resources/filter s [get]
//@Success 200 {object} ctx.JSONResult{result=resps.OrgEnvAndProviderResp}
func (Project) SearchProjectResourcesFilters(c *ctx.GinRequest) {
	form := forms.SearchProjectResourceForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchProjectResourcesFilters(c.Service(), &form))
}

// ProjectStat 项目概览统计数据
// @Tags 项目
// @Summary 项目概览统计数据
// @Description 项目概览统计数据
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param form formData forms.ProjectStatForm true "parameter"
// @router /projects/{id}/statistics [get]
// @Success 200 {object} ctx.JSONResult{result=resps.ProjectStatResp}
func (Project) ProjectStat(c *ctx.GinRequest) {
	form := forms.ProjectStatForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ProjectStat(c.Service(), &form))
}
