package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Task struct {
	ctrl.BaseController
}

// Search 任务查询
// @Tags 任务
// @Summary 任务查询
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param form query forms.SearchTaskForm true "parameter"
// @router /tasks [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Task}}
func (Task) Search(c *ctx.GinRequestCtx) {
	form := forms.SearchTaskForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.SearchTask(c.ServiceCtx(), &form))
}

// Detail 任务信息详情
// @Tags 任务
// @Summary 任务信息详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param taskId path string true "任务ID"
// @router /tasks/{taskId} [get]
// @Success 200 {object} ctx.JSONResult{result=apps.taskDetailResp}
func (Task) Detail(c *ctx.GinRequestCtx) {
	form := forms.DetailTaskForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.TaskDetail(c.ServiceCtx(), form))
}

// CurrentTask 环境当前任务详情
// @Tags 环境
// @Summary 环境当前任务详情
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param envId path string true "环境ID"
// @router /envs/{envId}/tasks/current [get]
// @Success 200 {object} ctx.JSONResult{result=apps.taskDetailResp}
func (Task) CurrentTask(c *ctx.GinRequestCtx) {
	form := &forms.CurrentTaskForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CurrentTask(c.ServiceCtx(), form))
}

// FollowLogSse 当前任务实时日志
// @Tags 任务
// @Summary 当前任务实时日志
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param taskId path string true "任务ID"
// @router /tasks/{taskId}/log/sse [get]
// @Success 200 {object} ctx.JSONResult{result=apps.taskDetailResp}
func (Task) FollowLogSse(c *ctx.GinRequestCtx) {
	// TODO 待实现
	//defer c.SSEvent("end", "end")
	//
	//if err := apps.FollowTaskLog(c); err != nil {
	//	c.SSEvent("error", err.Error())
	//}
}

// TaskStateListSearch
// @Tags 任务详情State List
// @Summary 任务详情State List
// @Accept application/json
// @Param taskGuid formData int true "任务guid"
// @router /template/state_list [get]
func (Task) TaskStateListSearch(c *ctx.GinRequestCtx) {
	// TODO 待实现
	//form := &forms.TaskStateListForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.TaskStateList(c.ServiceCtx(), form))
}

// TaskApprove 审批执行计划
// @Tags 任务
// @Summary 审批执行计划
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param taskId path string true "任务ID"
// @Param form formData forms.ApproveTaskForm true "parameter"
// @router /tasks/{taskId}/approve [post]
// @Success 200 {object} ctx.JSONResult{result=apps.taskDetailResp}
func (Task) TaskApprove(c *ctx.GinRequestCtx) {
	form := &forms.ApproveTaskForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.ApproveTask(c.ServiceCtx(), form))
}

// Log 任务日志
// @Tags 任务
// @Summary 任务日志
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param taskId path string true "任务ID"
// @router /tasks/{taskId}/log [get]
// @Success 200 {object} ctx.JSONResult{result=apps.taskDetailResp}
func (Task) Log(c *ctx.GinRequestCtx) {
	// TODO: 待实现
	//form := forms.DetailTaskForm{}
	//if err := c.Bind(&form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.TaskDetail(c.ServiceCtx(), form))
}

// Output Terraform Output
// @Tags 任务
// @Summary Terraform Output
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param Iac-Project-Id header string true "项目ID"
// @Param taskId path string true "任务ID"
// @router /tasks/{taskId}/output [get]
// @Success 200 {object} ctx.JSONResult{result=apps.taskDetailResp}
func (Task) Output(c *ctx.GinRequestCtx) {
	// TODO: 待实现
	//form := forms.DetailTaskForm{}
	//if err := c.Bind(&form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.TaskDetail(c.ServiceCtx(), form))
}
