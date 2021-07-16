package handlers

import (
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
)

type Key struct {
	ctrl.BaseController
}

// Create 创建密钥
// @Summary 创建密钥
// @Tags 密钥
// @Accept multipart/form-data
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param data formData forms.CreateKeyForm true "密钥信息"
// @Router /keys [post]
// @Success 200 {object} ctx.JSONResult{result=models.Key}
func (Key) Create(c *ctx.GinRequestCtx) {
	//form := &forms.CreateKeyForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.CreateKey(c.ServiceCtx(), form))
}

// Search 查询密钥
// @Summary 查询密钥
// @Tags 密钥
// @Accept application/x-www-form-urlencoded
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param data query forms.SearchKeyForm true "密钥查询参数"
// @Router /keys [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Key}}
func (Key) Search(c *ctx.GinRequestCtx) {
	//form := &forms.SearchKeyForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.SearchKey(c.ServiceCtx(), form))
}

// Update 修改密钥信息
// @Summary 修改密钥信息
// @Tags 密钥
// @Accept multipart/form-data
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param keyId path string true "密钥ID"
// @Param data formData forms.UpdateKeyForm true "密钥信息"
// @Router /keys/{keyId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.Key}
func (Key) Update(c *ctx.GinRequestCtx) {
	//form := &forms.UpdateKeyForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.UpdateKey(c.ServiceCtx(), form))
}

// Delete 删除密钥
// @Summary 删除密钥
// @Tags 密钥
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param keyId path string true "密钥ID"
// @Router /keys/{keyId} [delete]
// @Success 200
func (Key) Delete(c *ctx.GinRequestCtx) {
	//form := &forms.DeleteKeyForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.DeleteKey(c.ServiceCtx(), form))
}

// Detail 密钥详情
// @Summary 密钥详情
// @Tags 密钥
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param keyId path string true "密钥ID"
// @Router /keys/{keyId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.Key}
func (Key) Detail(c *ctx.GinRequestCtx) {
	//form := &forms.DetailKeyForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.DetailKey(c.ServiceCtx(), form))
}
