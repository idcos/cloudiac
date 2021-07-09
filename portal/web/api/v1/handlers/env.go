package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
)

type Env struct {
	ctrl.BaseController
}

// Create 创建环境
// @Tags 环境
// @Summary 创建环境
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param form formData forms.CreateEnvForm true "parameter"
// @router /envs [post]
// @Success 200 {object} ctx.JSONResult{result=models.Env}
func (Env) Create(c *ctx.GinRequestCtx) {
	form := forms.CreateEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.CreateEnv(c.ServiceCtx(), &form))
}

// Search 环境查询
// @Tags 环境
// @Summary 环境查询
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param form query forms.SearchEnvForm true "parameter"
// @router /envs [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.EnvDetail}}
func (Env) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchEnv(c.ServiceCtx(), &form))
}

// Update 环境编辑
// @Tags 环境
// @Summary 环境信息编辑
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param form formData forms.UpdateEnvForm true "parameter"
// @Param envId path string true "环境ID"
// @router /envs/{envId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.Env}
func (Env) Update(c *ctx.GinRequestCtx) {
	form := forms.UpdateEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateEnv(c.ServiceCtx(), &form))
}

// Detail 环境信息详情
// @Tags 环境
// @Summary 环境信息详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @router /envs/{envId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.EnvDetail}
func (Env) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.EnvDetail(c.ServiceCtx(), form))
}

// Archive 环境归档
// @Tags 环境
// @Summary 环境归档
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param form formData forms.ArchiveEnvForm true "parameter"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/archive [put]
// @Success 200 {object} ctx.JSONResult{result=models.EnvDetail}
func (Env) Archive(c *ctx.GinRequestCtx) {
	form := forms.UpdateEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateEnv(c.ServiceCtx(), &form))
}

// Deploy 环境重新部署
// @Tags 环境
// @Summary 环境重新部署
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param form formData forms.DeployEnvForm true "parameter"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/deploy [post]
// @Success 200 {object} ctx.JSONResult{result=models.EnvDetail}
func (Env) Deploy(c *ctx.GinRequestCtx) {
	form := forms.DeployEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.EnvDeploy(c.ServiceCtx(), &form))
}

// Destroy 销毁环境资源
// @Tags 环境
// @Summary 销毁环境资源
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param form formData forms.DestroyEnvForm true "parameter"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/destroy [post]
// @Success 200 {object} ctx.JSONResult{result=models.EnvDetail}
func (Env) Destroy(c *ctx.GinRequestCtx) {
	form := forms.DeployEnvForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	form.TaskType = models.TaskTypeDestroy
	c.JSONResult(apps.EnvDeploy(c.ServiceCtx(), &form))
}

// SearchResources 获取环境资源列表
// @Tags 环境
// @Summary 获取环境资源列表
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param form formData forms.SearchEnvResourceForm true "parameter"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/resources [post]
// @Success 200 {object} ctx.JSONResult{result=models.EnvRes}
func (Env) SearchResources(c *ctx.GinRequestCtx) {
	// TODO: 待实现
	//form := forms.SearchEnvResourceForm{}
	//if err := c.Bind(&form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.UpdateEnv(c.ServiceCtx(), &form))
}
