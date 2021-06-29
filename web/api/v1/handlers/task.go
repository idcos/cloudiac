package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type Task struct {
	ctrl.BaseController
}

// Detail 查询作业详情
// @Summary 查询作业详情
// @Description 查询作业详情
// @Tags 作业
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param taskId query int true "作业id"
// @Success 200 {object} apps.DetailTaskResp
// @Router /task/detail [get]
func (Task) Detail(c *ctx.GinRequestCtx) {
	form := &forms.DetailTaskForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailTask(c.ServiceCtx(), form))
}

// Create 创建作业
// @Summary 创建作业
// @Description 创建作业
// @Tags 作业
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.CreateTaskForm true "作业信息"
// @Success 200 {object} models.Task
// @Router /task/create [post]
func (Task) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateTaskForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateTask(c.ServiceCtx(), form))
}

// Search 查询作业列表
// @Summary 查询作业列表
// @Description 查询作业列表
// @Tags 作业
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param q query string false "模糊搜索"
// @Param status query string false "作业状态"
// @Param templateId query string true "模板id"
// @Success 200 {object} apps.SearchTaskResp
// @Router /task/search [get]
func (Task) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchTaskForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchTask(c.ServiceCtx(), form))
}

// LastTask 查询最后一次作业
// @Summary 查询最后一次作业
// @Description 查询最后一次作业
// @Tags 作业
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param templateId query string true "模板id"
// @Success 200 {object} apps.LastTaskResp
// @Router /task/last [get]
func (Task) LastTask(c *ctx.GinRequestCtx) {
	form := &forms.LastTaskForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.LastTask(c.ServiceCtx(), form))
}

func (Task) FollowLogSse(c *ctx.GinRequestCtx) {
	defer c.SSEvent("end", "end")

	if err := apps.FollowTaskLog(c); err != nil {
		c.SSEvent("error", err.Error())
	}
}

// TaskStateListSearch 作业详情State List
// @Summary 作业详情State List
// @Description 作业详情State List
// @Tags 作业
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param taskGuid query string true "作业guid"
// @Success 200 {object} []string
// @Router /template/state_list [get]
func (Task) TaskStateListSearch(c *ctx.GinRequestCtx) {
	form := &forms.TaskStateListForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.TaskStateList(c.ServiceCtx(), form))
}
