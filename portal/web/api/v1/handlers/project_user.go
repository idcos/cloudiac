package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type ProjectUser struct {
	ctrl.BaseController
}

// Create 创建项目
// @Summary 创建项目
// @Description 创建项目
// @Tags 项目
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param IaC-Project-Id header string true "项目id"
// @Param json body forms.CreateProjectUserForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=models.Project}
// @Router /projects/users [post]
func (ProjectUser) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateProjectUserForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateProjectUser(c.ServiceCtx(), form))
}

// Search 查询项目列表
// @Summary 查询项目列表
// @Description 查询项目列表
// @Tags 项目
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param IaC-Project-Id header string true "项目id"
// @Param form query forms.SearchProjectOrgUserForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Project}}
// @Router /projects/users [get]
func (ProjectUser) Search(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.SearchProjectUser(c.ServiceCtx()))
}

// Update 修改项目用户授权信息
// @Summary 修改项目用户授权信息
// @Description 修改项目用户授权信息
// @Tags 项目
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param IaC-Project-Id header string true "项目id"
// @Param id query string true "用户项目id"
// @Param request body forms.UpdateProjectUserForm true "用户授权"
// @Success 200 {object} ctx.JSONResult{result=models.Project}
// @Router /projects/users/{id}  [put]
func (ProjectUser) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateProjectUserForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateProjectUser(c.ServiceCtx(), form))
}

// Delete 删除项目信息
// @Summary 删除项目信息
// @Description 删除项目信息
// @Tags 项目
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param IaC-Project-Id header string true "项目id"
// @Param id query string true "用户项目id"
// @Success 200
// @Router /projects/users/{id} [delete]
func (ProjectUser) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteProjectOrgUserForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteProjectUser(c.ServiceCtx(), form))
}
