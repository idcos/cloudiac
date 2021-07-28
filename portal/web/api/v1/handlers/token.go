package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Token struct {
	ctrl.GinController
}

// Create 创建token
// @Summary 创建token
// @Description 创建token
// @Tags Token
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param data body forms.CreateTokenForm true "ApiToken信息"
// @Success 200 {object} ctx.JSONResult{result=models.Token}
// @Router /tokens [post]
func (Token) Create(c *ctx.GinRequest) {
	form := &forms.CreateTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateToken(c.Service(), form))
}

// Search 查询token
// @Summary 查询token
// @Description 查询token
// @Tags Token
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param q query string false "模糊搜索"
// @Param status query string false "ApiToken状态"
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Token}}
// @Router /tokens [get]
func (Token) Search(c *ctx.GinRequest) {
	form := &forms.SearchTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchToken(c.Service(), form))
}

// Update 修改token信息
// @Summary 修改token信息
// @Description 修改token信息
// @Tags Token
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param tokenId path string true "TokenID"
// @Param data body forms.UpdateTokenForm true "ApiToken信息"
// @Success 200
// @Router /tokens/{tokenId} [put]
func (Token) Update(c *ctx.GinRequest) {
	form := &forms.UpdateTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateToken(c.Service(), form))
}

// Delete 删除Token账号
// @Summary 删除Token账号
// @Description 删除Token账号
// @Tags Token
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param tokenId path string true "TokenID"
// @Param data body forms.DeleteTokenForm true "DeleteTokenForm信息"
// @Success 200
// @Router /tokens/{tokenId} [delete]
func (Token) Delete(c *ctx.GinRequest) {
	form := &forms.DeleteTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteToken(c.Service(), form))
}

// DetailTriggerToken 触发器token详情
// @Summary 触发器token详情
// @Description 触发器token详情
// @Tags Token
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param data body forms.DetailTriggerTokenForm true "DeleteTokenForm信息"
// @Success 200 {object} ctx.JSONResult{result=models.Token}
// @Router /tokens/trigger [get]
func (Token) DetailTriggerToken(c *ctx.GinRequest) {
	form := &forms.DetailTriggerTokenForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailTriggerToken(c.Service(), form))
}

func ApiTriggerHandler(c *ctx.GinRequest) {
	form := forms.ApiTriggerHandler{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ApiTriggerHandler(c.Service(), form))
}
