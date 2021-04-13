package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/libs/page"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
)

func CreateTaskComment(c *ctx.ServiceCtx, form *forms.CreateTaskCommentForm) (interface{}, e.Error) {
	return services.CreateTaskComment(c.DB().Debug(), models.TaskComment{
		TaskId:    form.TaskId,
		Creator:   c.Username,
		CreatorId: c.UserId,
		Comment:   form.Comment,
	})
}

func SearchTaskComment(c *ctx.ServiceCtx, form *forms.SearchTaskCommentForm) (interface{}, e.Error) {
	query := services.SearchTaskComment(c.DB(), form.TaskId)
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
