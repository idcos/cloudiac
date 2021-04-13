package forms

type CreateTaskCommentForm struct {
	BaseForm

	TaskId  uint   `json:"taskId" form:"taskId" binding:"required"`
	Comment string `json:"comment" form:"comment" binding:"required"`
}

type SearchTaskCommentForm struct {
	BaseForm
	TaskId uint `json:"taskId" form:"taskId" `
}
