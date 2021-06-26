package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctrl"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

type TaskComment struct {
	ctrl.BaseController
}

// Create 创建作业评论
// @Summary 创建作业评论
// @Description 创建作业评论
// @Tags 作业
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param data body forms.CreateTaskCommentForm true "作业评论信息"
// @Success 200 {object} models.TaskComment
// @Router /task/comment/create [post]
func (TaskComment) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateTaskCommentForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateTaskComment(c.ServiceCtx(), form))
}

// Search 查询作业评论列表
// @Summary 查询作业评论列表
// @Description 查询作业评论列表
// @Tags 作业
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param taskId query int true "作业id"
// @Success 200 {object} models.TaskComment
// @Router /task/comment/search [get]
func (TaskComment) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchTaskCommentForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchTaskComment(c.ServiceCtx(), form))
}
