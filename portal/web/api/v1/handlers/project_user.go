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

// Create 创建项目下的用户
// @Summary 创建项目下的用户
// @Description 创建项目下的用户
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

// Search 查询组织下用户列表（不包含已经加入项目的用户）
// @Summary 查询组织下用户列表（不包含已经加入项目的用户）
// @Description 查询组织下用户列表（不包含已经加入项目的用户）
// @Tags 项目
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param IaC-Project-Id header string true "项目id"
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

// Delete 移除项目下的用户
// @Summary 移除项目下的用户
// @Description 移除项目下的用户
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

// SearchProjectAuthorizationUser 项目下授权的用户
// @Summary  项目下授权的用户
// @Description  项目下授权的用户
// @Tags 项目
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param IaC-Project-Id header string true "项目id"
// @Param request body forms.SearchProjectAuthorizationUserForm true "用户授权"
// @Success 200
// @Router /projects/authorization/users [get]
func (ProjectUser) SearchProjectAuthorizationUser(c *ctx.GinRequestCtx) {
	form := &forms.SearchProjectAuthorizationUserForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchProjectAuthorizationUser(c.ServiceCtx(), form))
}
