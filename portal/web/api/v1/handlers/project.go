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

func (Project) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchProject(c.ServiceCtx(), form))
}

func (Project) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateProject(c.ServiceCtx(), form))
}

func (Project) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteProject(c.ServiceCtx(), form))
}

func (Project) Detail(c *ctx.GinRequestCtx) {
	form := &forms.DetailProjectForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailProject(c.ServiceCtx(), form))
}
