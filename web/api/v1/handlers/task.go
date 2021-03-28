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

func (Task) Detail(c *ctx.GinRequestCtx)  {
	form := &forms.DetailTaskForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DetailTask(c.ServiceCtx(), form))
}

func (Task) Create(c *ctx.GinRequestCtx)  {
	form := &forms.CreateTaskForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateTask(c.ServiceCtx(), form))
}

func (Task) Search(c *ctx.GinRequestCtx)  {
	form := &forms.SearchTaskForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchTask(c.ServiceCtx(), form))
}
