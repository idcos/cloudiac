package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Policy struct {
	ctrl.GinController
}

// Create 创建策略
// @Summary 创建策略
// @Description 创建策略
// @Tags 策略
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param json body forms.CreatePolicyForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=models.Policy}
// @Router /policies [post]
func (Policy) Create(c *ctx.GinRequest) {
	form := &forms.CreatePolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreatePolicy(c.Service(), form))
}

// Parse TODO: 解析云模板
// @Summary 解析云模板
// @Description 解析云模板
// @Tags 策略
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param json body forms.CreatePolicyForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=models.Policy}
// @Router /policies/parse [post]
func (Policy) Parse(c *ctx.GinRequest) {
	// TODO
	//form := &forms.CreatePolicyForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.CreatePolicy(c.Service(), form))
}

// ScanTemplate 运行云模板策略扫描
// @Summary 运行云模板策略扫描
// @Description 运行云模板策略扫描
// @Tags 策略
// @Accept  json
// @Produce  json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织id"
// @Param templateId path string true "模板ID"
// @Param json body forms.ScanTemplateForm true "parameter"
// @Success 200 {object}  ctx.JSONResult{result=models.ScanTask}
// @Router /policies/templates/{templateId}/scan [post]
func (Policy) ScanTemplate(c *ctx.GinRequest) {
	form := &forms.ScanTemplateForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.ScanTemplate(c.Service(), form))
}

// Search 查询策略列表
// @Tags 策略
// @Summary 查询策略列表
// @Description 查询策略列表
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param q query string false "模糊搜索"
// @Param severity query string false "严重性"
// @Param groupId query string false "策略组Id"
// @Router /policies [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Policy}}
func (Policy) Search(c *ctx.GinRequest) {
	form := &forms.SearchPolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicy(c.Service(), form))
}

// Update 修改策略
// @Tags 策略
// @Summary 修改策略
// @Description 修改策略
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param json body forms.UpdatePolicyForm true "parameter"
// @Param policiesId path string true "策略Id"
// @Router /policies/{policiesId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.Policy}
func (Policy) Update(c *ctx.GinRequest) {
	form := &forms.UpdatePolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdatePolicy(c.Service(), form))
}

// Delete 删除策略
// @Tags 策略
// @Summary 删除策略
// @Description 删除策略
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param policiesId path string true "策略Id"
// @Router /policies/{policiesId} [delete]
// @Success 200 {object} ctx.JSONResult
func (Policy) Delete(c *ctx.GinRequest) {
	form := &forms.DeletePolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeletePolicy(c.Service(), form))
}

// Detail 策略详情
// @Tags 策略
// @Summary Detail
// @Description 策略详情
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param policiesId path string true "策略Id"
// @Router /policies/{policiesId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.Policy}
func (Policy) Detail(c *ctx.GinRequest) {
	form := &forms.DetailPolicyForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailPolicy(c.Service(), form))
}

// SearchPolicyTpl 查询云模板策略配置
// @Tags 策略
// @Summary Search
// @Description 查询云模板策略配置
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param q query string false "模糊搜索"
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Router /policies/templates [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]apps.RespPolicyTpl}}
func (Policy) SearchPolicyTpl(c *ctx.GinRequest) {
	form := &forms.SearchPolicyTplForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicyTpl(c.Service(), form))
}

// UpdatePolicyTpl 修改云模板与策略组关联
// @Tags 策略
// @Summary Update
// @Description 修改云模板与策略组关联
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param json body forms.UpdatePolicyTplForm true "parameter"
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Router /policies/templates [put]
// @Success 200 {object} ctx.JSONResult
func (Policy) UpdatePolicyTpl(c *ctx.GinRequest) {
	form := &forms.UpdatePolicyTplForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdatePolicyTpl(c.Service(), form))
}

// DetailPolicyTpl 云模板策略详情
// @Tags 策略
// @Summary Detail
// @Description 云模板策略详情
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param tplId path string true "模板id"
// @Router /policies/templates/{tplId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.Policy}
func (Policy) DetailPolicyTpl(c *ctx.GinRequest) {
	form := &forms.DetailPolicyTplForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailPolicyTpl(c.Service(), form))
}

// SearchPolicyEnv 查询环境策略配置
// @Tags 策略
// @Summary Search
// @Description 查询环境策略配置
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param q query string false "模糊搜索"
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Router /policies/envs [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]apps.RespPolicyEnv}}
func (Policy) SearchPolicyEnv(c *ctx.GinRequest) {
	form := &forms.SearchPolicyEnvForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchPolicyEnv(c.Service(), form))
}

// UpdatePolicyEnv 修改环境与策略组关联
// @Tags 策略
// @Summary Update
// @Description 修改环境与策略组关联
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param json body forms.UpdatePolicyEnvForm true "parameter"
// @Router /policies/envs [put]
// @Success 200 {object} ctx.JSONResult
func (Policy) UpdatePolicyEnv(c *ctx.GinRequest) {
	form := &forms.UpdatePolicyEnvForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdatePolicyEnv(c.Service(), form))
}

// DetailPolicyEnv 环境策略详情
// @Tags 策略
// @Summary Detail
// @Description 环境策略详情
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param envId path string true "环境id"
// @Router /policies/envs/{envId} [get]
// @Success 200 {object} ctx.JSONResult{result=models.Policy}
func (Policy) DetailPolicyEnv(c *ctx.GinRequest) {
	form := &forms.DetailPolicyEnvForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailPolicyEnv(c.Service(), form))
}

// PolicyError 策略详情-错误
// @Tags 策略
// @Summary Detail
// @Description 策略详情-错误
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param policyId path string true "策略id"
// @Router /policies/{policyId}/error [get]
// @Success 200 {object} ctx.JSONResult{result=models.Policy}
func (Policy) PolicyError(c *ctx.GinRequest) {
	form := &forms.PolicyErrorForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.PolicyError(c.Service(), form))
}

// PolicyReference 策略详情-参考
// @Tags 策略
// @Summary Detail
// @Description 策略详情-参考
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param policyId path string true "策略id"
// @Router /policies/{policyId}/reference [get]
// @Success 200 {object} ctx.JSONResult{result=models.Policy}
func (Policy) PolicyReference(c *ctx.GinRequest) {
	form := &forms.PolicyReferenceForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.PolicyReference(c.Service(), form))
}

// PolicyRepo 策略详情-报表
// @Tags 策略
// @Summary Detail
// @Description 策略详情-报表
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param IaC-Org-Id header string true "组织ID"
// @Param IaC-Project-Id header string false "项目ID"
// @Param policyId path string true "策略id"
// @Router /policies/{policyId}/report [get]
// @Success 200 {object} ctx.JSONResult{result=models.Policy}
func (Policy) PolicyRepo(c *ctx.GinRequest) {
	form := &forms.PolicyRepoForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.PolicyRepo(c.Service(), form))
}
