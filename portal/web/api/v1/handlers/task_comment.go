package handlers

import (
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
)

type TaskComment struct {
	ctrl.BaseController
}

func (TaskComment) Create(c *ctx.GinRequestCtx) {
	// TODO 待实现
	//form := &forms.CreateTaskCommentForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.CreateTaskComment(c.ServiceCtx(), form))
}

func (TaskComment) Search(c *ctx.GinRequestCtx) {
	// TODO 待实现
	//form := &forms.SearchTaskCommentForm{}
	//if err := c.Bind(form); err != nil {
	//	return
	//}
	//c.JSONResult(apps.SearchTaskComment(c.ServiceCtx(), form))
}
