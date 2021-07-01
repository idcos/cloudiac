package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Project struct {
	ctrl.BaseController
}

// Create 创建项目
// @Summary 创建项目
// @Description 创建项目
// @Tags 项目
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param name body string true "项目名称"
// @Param description body string false "项目描述"
// @Param userAuthorization body forms.UserAuthorization false "用户授权"
// @Success 200 {object} models.Project
// @Router /project/create [post]
func (Project) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateProject(c.ServiceCtx(), form))

}

// Search 查询项目列表
// @Summary 查询项目列表
// @Description 查询项目列表
// @Tags 项目
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param q query string false "模糊搜索"
// @Param currentPage query int false "分页页码"
// @Param pageSize query int false "分页页数"
// @Success 200 {object} models.Project
// @Router /project/search [get]
func (Project) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchProject(c.ServiceCtx(), form))
}

// Update 修改项目信息
// @Summary 修改项目信息
// @Description 修改项目信息
// @Tags 项目
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param id body string true "项目id"
// @Param name body string false "项目名称"
// @Param description body string false "项目描述"
// @Param userAuthorization body []forms.UserAuthorization false "用户授权"
// @Success 200 {object} models.Project
// @Router /project/update [put]
func (Project) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateProject(c.ServiceCtx(), form))
}

// Delete 删除项目信息
// @Summary 删除项目信息
// @Description 删除项目信息
// @Tags 项目
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param id body string true "项目id"
// @Success 200
// @Router /project/delete [delete]
func (Project) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteProject(c.ServiceCtx(), form))
}

// Detail 查询项目详情
// @Summary 查询项目详情
// @Description 查询项目详情
// @Tags 项目
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param id query string true "项目id"
// @Success 200 {object} models.Project
// @Router /project/detail [get]
func (Project) Detail(c *ctx.GinRequestCtx) {
	form := &forms.DetailProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailProject(c.ServiceCtx(), form))
}
