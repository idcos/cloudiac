package handlers

import (
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
)

type Task struct {
	ctrl.BaseController
}

func (Task) Detail(c *ctx.GinRequestCtx) {
	// TODO 待实现
	//form := &forms.DetailTaskForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.DetailTask(c.ServiceCtx(), form))
}

func (Task) Create(c *ctx.GinRequestCtx) {
	// TODO 待实现
	//form := &forms.CreateTaskForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.CreateTask(c.ServiceCtx(), form))
}

func (Task) Search(c *ctx.GinRequestCtx) {
	// TODO 待实现
	//form := &forms.SearchTaskForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.SearchTask(c.ServiceCtx(), form))
}

func (Task) LastTask(c *ctx.GinRequestCtx) {
	// TODO 待实现
	//form := &forms.LastTaskForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.LastTask(c.ServiceCtx(), form))
}

func (Task) FollowLogSse(c *ctx.GinRequestCtx) {
	// TODO 待实现
	//defer c.SSEvent("end", "end")
	//
	//if err := apps.FollowTaskLog(c); err != nil {
	//	c.SSEvent("error", err.Error())
	//}
}


func (Task) TaskStateListSearch(c *ctx.GinRequestCtx) {
	// TODO 待实现
	//form := &forms.TaskStateListForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.TaskStateList(c.ServiceCtx(), form))
}
