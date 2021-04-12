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

func (TaskComment) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateTaskCommentForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateTaskComment(c.ServiceCtx(), form))
}

func (TaskComment) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchTaskCommentForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchTaskComment(c.ServiceCtx(), form))
}
