// Copyright 2021 CloudJ Company Limited. All rights reserved.

package apps

import (
	"cloudiac/portal/consts/e"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/libs/page"
	"cloudiac/portal/models"
	"cloudiac/portal/models/forms"
	"cloudiac/portal/services"
)

func CreateTaskComment(c *ctx.ServiceContext, form *forms.CreateTaskCommentForm) (interface{}, e.Error) {
	return services.CreateTaskComment(c.DB(), models.TaskComment{
		TaskId:    form.Id,
		Creator:   c.Username,
		CreatorId: c.UserId,
		Comment:   form.Comment,
	})
}

func SearchTaskComment(c *ctx.ServiceContext, form *forms.SearchTaskCommentForm) (interface{}, e.Error) {
	query := services.SearchTaskComment(c.DB(), form.Id)
	p := page.New(form.CurrentPage(), form.PageSize(), query)
	taskComment := make([]*models.TaskComment, 0)
	if err := p.Scan(&taskComment); err != nil {
		return nil, e.New(e.DBError, err)
	}

	return page.PageResp{
		Total:    p.MustTotal(),
		PageSize: p.Size,
		List:     taskComment,
	}, nil
}
