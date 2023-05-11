package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Tag struct {
	ctrl.GinController
}

// Create 环境标签新建
// @Tags 环境
// @Summary 环境标签新建
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param tags formData forms.EnvUnLockConfirmForm true "部署参数"
// @router /envs/{envId}/tags [post]
// @Success 200 {object} ctx.JSONResult{result=resps.EnvUnLockConfirmResp}
func (Tag) Create(c *ctx.GinRequest) {
	form := forms.CreateTagForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.CreateTag(c.Service(), &form))
}

// Search 查询云模板列表
// @Tags 云模板
// @Summary 查询云模板列表
// @Accept multipart/x-www-form-urlencoded
// @Security AuthToken
// @Description 查询云模板列表
// @Tags 云模板
// @Accept  json
// @Produce  json
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param form query forms.SearchTemplateForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.SearchTemplateResp}}
// @Router /templates [get]
func (Tag) Search(c *ctx.GinRequest) {
	form := forms.SearchTemplateForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchTag(c.Service(), &form))
}

// Delete 环境标签删除
// @Tags 环境
// @Summary 环境标签删除
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param tags formData forms.EnvUnLockConfirmForm true "部署参数"
// @router /envs/{envId}/tags/{tagId} [delete]
// @Success 200 {object} ctx.JSONResult{result=resps.EnvUnLockConfirmResp}
func (Tag) Delete(c *ctx.GinRequest) {
	form := forms.DeleteTagForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteTag(c.Service(), &form))
}

// Update 环境标签修改
// @Tags 环境
// @Summary 环境标签修改
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @Param tags formData forms.EnvUnLockConfirmForm true "部署参数"
// @router /envs/{envId}/tags/{tagId} [put]
// @Success 200 {object} ctx.JSONResult{result=resps.EnvUnLockConfirmResp}
func (Tag) Update(c *ctx.GinRequest) {
	form := forms.UpdateTagForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateTag(c.Service(), &form))
}
