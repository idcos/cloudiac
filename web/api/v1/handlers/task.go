package handlers

import (
	"cloudiac/apps"
	"cloudiac/libs/ctx"
	"cloudiac/models/forms"
)

func TaskDetail(c *ctx.GinRequestCtx)  {
	form := &forms.DetailTaskForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.TaskDetail(c.ServiceCtx(),form))
}

func TaskCreate(c *ctx.GinRequestCtx)  {
	form := &forms.CreateTaskForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.TaskCreate(c.ServiceCtx(),form))
}

