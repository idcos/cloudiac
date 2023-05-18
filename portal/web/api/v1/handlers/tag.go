// Copyright (c) 2015-2023 CloudJ Technology Co., Ltd.

package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/consts"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
	"fmt"
)

type Tag struct {
	ctrl.GinController
}

// Create 创建标签
// @Tags 标签
// @Summary 创建标签
// @Accept multipart/form-data
// @Accept json
// @Security AuthToken
// @Produce json
// @Param IaC-Org-Id header string true "组织ID"
// @Param json body forms.CreateTagForm true "parameter"
// @Router /tags [post]
// @Success 200 {object} ctx.JSONResult{result=models.TagValue}
func (Tag) Create(c *ctx.GinRequest) {
	form := &forms.CreateTagForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateTag(c.Service(), form))
}

// Search 查询标签列表
// @Tags 标签
// @Summary 查询标签列表
// @Accept multipart/x-www-form-urlencoded
// @Security AuthToken
// @Description 查询标签列表
// @Tags 标签
// @Accept  json
// @Produce  json
// @Param IaC-Org-Id header string true "组织ID"
// @Param form query forms.SearchTagsForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.RespTag}}
// @Router /tags [get]
func (Tag) Search(c *ctx.GinRequest) {
	form := forms.SearchTagsForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchTag(c.Service(), &form))
}

// Update 修改标签信息
// @Tags 标签
// @Summary 修改标签信息
// @Description 修改标签信息
// @Security AuthToken
// @Accept  json
// @Produce  json
// @Param tagId path string true "标签ID"
// @Param data body forms.UpdateTemplateForm true "云模板信息"
// @Success 200 {object} ctx.JSONResult{result=models.TagValue}
// @Router /tags [put]
func (Tag) Update(c *ctx.GinRequest) {
	form := forms.UpdateTagsForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateTag(c.Service(), &form))
}

// Delete 删除标签
// @Summary 删除标签
// @Tags 标签
// @Description 删除标签
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Router /tags [delete]
// @Success 200 {object} ctx.JSONResult
func (Tag) Delete(c *ctx.GinRequest) {
	form := forms.DeleteTagsForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteTag(c.Service(), &form))
}

// SearchEnvTag 查询环境标签
// @Summary 查询环境标签
// @Tags 标签
// @Description 查询环境标签
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param id path string true "环境id"
// @Router /envs/{id}/tags [get]
// @Param form query forms.SearchTagsForm true "parameter"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]resps.RespTag}}
func SearchEnvTag(c *ctx.GinRequest) {
	form := forms.SearchEnvTagsForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	fmt.Println(form.Id,"form.Id")
	c.JSONResult(apps.SearchTag(c.Service(), &forms.SearchTagsForm{
		PageForm:   form.PageForm,
		Q:          form.Q,
		ObjectType: consts.Env,
		ObjectId:   form.Id,
	}))
}
